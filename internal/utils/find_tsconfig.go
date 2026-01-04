package utils

import (
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/typescript-eslint/tsgolint/internal/collections"
)

type TsConfigResolver struct {
	fs                        vfs.FS
	currentDirectory          string
	configFileRegistryBuilder *project.ConfigFileRegistryBuilder
}

func NewTsConfigResolver(fs vfs.FS, currentDirectory string) *TsConfigResolver {
	return &TsConfigResolver{
		fs:               fs,
		currentDirectory: currentDirectory,
		configFileRegistryBuilder: project.NewConfigFileRegistryBuilder(
			project.TsGoLintNewSnapshotFSBuilder(fs, currentDirectory), &project.ConfigFileRegistry{}, project.NewExtendedConfigCache(), &project.SessionOptions{
				CurrentDirectory: currentDirectory,
			}, nil),
	}
}

// Finds the tsconfig.json that governs the given file
// Reference: `findOrCreateDefaultConfiguredProjectForOpenScriptInfo` typescript-go/internal/project/projectcollectionbuilder.go:629-671
func (r *TsConfigResolver) FindTsconfigForFile(filePath string, skipSearchInDirectoryOfFile bool) (configPath string, found bool) {
	configFileName := r.configFileRegistryBuilder.ComputeConfigFileName(filePath, skipSearchInDirectoryOfFile, nil)

	if configFileName == "" {
		return "", false
	}

	normalizedPath := tspath.ToPath(filePath, r.currentDirectory, r.fs.UseCaseSensitiveFileNames())

	// Search through the config and its references
	// This corresponds to findOrCreateDefaultConfiguredProjectWorker
	result := r.findConfigWithReferences(filePath, normalizedPath, configFileName, nil, nil)

	if result.configFileName != "" {
		return result.configFileName, true
	}

	return "", false
}

// Reference: `searchResult`: typescript-go/internal/project/projectcollectionbuilder.go:461-465
type configSearchResult struct {
	configFileName string
}

// Reference: `searchNode`: typescript-go/internal/project/projectcollectionbuilder.go:467-471
type searchNode struct {
	configFileName string
}

// Reference: `findOrCreateDefaultConfiguredProjectWorker`: typescript-go/internal/project/projectcollectionbuilder.go:480-627
func (r *TsConfigResolver) findConfigWithReferences(
	fileName string,
	path tspath.Path,
	configFileName string,
	visited *collections.SyncSet[searchNode],
	fallback *configSearchResult,
) configSearchResult {
	var configs collections.SyncMap[tspath.Path, *tsoptions.ParsedCommandLine]
	if visited == nil {
		visited = &collections.SyncSet[searchNode]{}
	}

	search := BreadthFirstSearch(
		searchNode{configFileName: configFileName},
		func(node searchNode) []searchNode {
			if config, ok := configs.Load(r.toPath(node.configFileName)); ok && len(config.ProjectReferences()) > 0 {
				references := config.ResolvedProjectReferencePaths()
				return Map(references, func(configFileName string) searchNode {
					return searchNode{configFileName: configFileName}
				})
			}
			return nil
		},
		func(node searchNode) (isResult bool, stop bool) {
			configFilePath := r.toPath(node.configFileName)

			config := r.configFileRegistryBuilder.FindOrAcquireConfigForFile(
				node.configFileName, configFilePath, path, project.ProjectLoadKindCreate, nil,
			)
			if config == nil {
				return false, false
			}
			configs.Store(configFilePath, config)
			if len(config.FileNames()) == 0 {
				return false, false
			}
			if config.CompilerOptions().Composite == core.TSTrue {
				// For composite projects, we can get an early negative result.
				// !!! what about declaration files in node_modules? wouldn't it be better to
				//     check project inclusion if the project is already loaded?
				if !config.PossiblyMatchesFileName(fileName) {
					return false, false
				}
			}

			if slices.ContainsFunc(config.FileNames(), func(file string) bool {
				// Fast checks:
				// 1) check if the strings happen to already be equal (subject to case sensitivity of FS)
				// 2) check if the base names are equal (subject to case sensitivity of FS)

				// If we're on a case-insensitive FS and the strings are equal, we can return true immediately,
				// no need to allocate and do any path conversions.
				if r.fs.UseCaseSensitiveFileNames() {
					if file == string(path) {
						return true
					}
				} else {
					if strings.EqualFold(file, string(path)) {
						return true
					}
				}

				// If the base names don't match, we can return false immediately.
				pathBaseName := filepath.Base(string(path))
				fileBaseName := filepath.Base(file)
				if r.fs.UseCaseSensitiveFileNames() {
					if fileBaseName != pathBaseName {
						return false
					}
				} else {
					if !strings.EqualFold(fileBaseName, pathBaseName) {
						return false
					}
				}

				// Finally, do a full path conversion and comparison (note: this allocates)
				return r.toPath(file) == path
			}) {
				return true, true
			}

			return false, false
		},
		BreadthFirstSearchOptions[searchNode]{
			Visited: visited,
			PreprocessLevel: func(level *BreadthFirstSearchLevel[searchNode]) {
				level.Range(func(node searchNode) bool {
					return true
				})
			},
		},
	)

	tsconfig := ""
	if len(search.Path) > 0 {
		tsconfig = search.Path[0].configFileName
	} else {
		tsconfig = ""
	}

	if search.Stopped {
		return configSearchResult{configFileName: tsconfig}
	}
	if tsconfig != "" {
		fallback = &configSearchResult{configFileName: tsconfig}
	}

	// Look for tsconfig.json files higher up the directory tree and do the same. This handles
	// the common case where a higher-level "solution" tsconfig.json contains all projects in a
	// workspace.
	if config, ok := configs.Load(r.toPath(configFileName)); ok && config.CompilerOptions().DisableSolutionSearching.IsTrue() {
		if fallback != nil {
			return *fallback
		}
	}

	if ancestorConfigName := r.configFileRegistryBuilder.GetAncestorConfigFileName(fileName, path, configFileName, nil); ancestorConfigName != "" {
		return r.findConfigWithReferences(
			fileName,
			path,
			ancestorConfigName,
			visited,
			fallback,
		)
	}
	if fallback != nil {
		return *fallback
	}

	return configSearchResult{configFileName: ""}
}

type ResolutionResult struct {
	file   string
	config string
}

func (r *TsConfigResolver) work(in <-chan string, out chan<- ResolutionResult) {
	for file := range in {
		config := r.configFileRegistryBuilder.ComputeConfigFileName(file, false, nil)
		if config == "" {
			out <- ResolutionResult{
				file:   file,
				config: config,
			}
			continue
		}

		fileNormalized := tspath.ToPath(file, r.currentDirectory, r.fs.UseCaseSensitiveFileNames())

		// Search through the config and its references
		result := r.findConfigWithReferences(file, fileNormalized, config, nil, nil)

		if result.configFileName != "" {
			out <- ResolutionResult{
				config: result.configFileName,
				file:   file,
			}
			continue
		}

		out <- ResolutionResult{
			config: result.configFileName,
			file:   file,
		}
	}
}

func (r *TsConfigResolver) FindTsConfigParallel(fileNames []string) map[string]string {
	in := make(chan string, len(fileNames))
	out := make(chan ResolutionResult, len(fileNames))

	numWorker := runtime.GOMAXPROCS(0)

	var wg sync.WaitGroup
	for range numWorker {
		wg.Go(func() {
			r.work(in, out)
		})
	}

	for i := range fileNames {
		in <- fileNames[i]
	}
	close(in)

	go func() {
		wg.Wait()
		close(out)
	}()

	res := make(map[string]string, len(fileNames))
	for result := range out {
		res[result.file] = result.config
	}

	return res
}

// Reference: `toPath`: typescript-go/internal/project/projectcollectionbuilder.go:687-689
func (b *TsConfigResolver) toPath(fileName string) tspath.Path {
	return tspath.ToPath(fileName, b.currentDirectory, b.fs.UseCaseSensitiveFileNames())
}

func (r *TsConfigResolver) FS() vfs.FS {
	return r.fs
}

func (r *TsConfigResolver) GetCurrentDirectory() string {
	return r.currentDirectory
}
