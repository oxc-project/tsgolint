package utils

import (
	"log"
	"slices"

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
			project.TsGoLintNewSnapshotFSBuilder(fs), &project.ConfigFileRegistry{}, &project.ExtendedConfigCache{}, &project.SessionOptions{
				CurrentDirectory: currentDirectory,
			}, nil),
	}
}

// Finds the tsconfig.json that governs the given file
// Reference: `findOrCreateDefaultConfiguredProjectForOpenScriptInfo` typescript-go/internal/project/projectcollectionbuilder.go:629-671
func (r *TsConfigResolver) FindTsconfigForFile(filePath string, skipSearchInDirectoryOfFile bool) (configPath string, found bool) {
	configFileName := r.configFileRegistryBuilder.ComputeConfigFileName(filePath, skipSearchInDirectoryOfFile, nil)

	log.Println("got tsconfig file name: " + configFileName)

	if configFileName == "" {
		return "", false
	}

	normalizedPath := tspath.ToPath(filePath, r.currentDirectory, r.fs.UseCaseSensitiveFileNames())

	// Search through the config and its references
	// This corresponds to findOrCreateDefaultConfiguredProjectWorker
	result := r.findConfigWithReferences(filePath, normalizedPath, configFileName, nil, nil)
	log.Println("found tsconfig file: " + result.configFileName)

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

	search := BreadthFirstSearchParallelEx(
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

			config := r.configFileRegistryBuilder.FindOrAcquireConfigForOpenFile(
				node.configFileName, configFilePath, path, project.ProjectLoadKindCreate, nil,
			)
			if config == nil {
				return false, false
			}
			if len(config.FileNames()) == 0 {
				return false, false
			}
			if config.CompilerOptions().Composite == core.TSTrue {
				// For composite projects, we can get an early negative result.
				// !!! what about declaration files in node_modules? wouldn't it be better to
				//     check project inclusion if the project is already loaded?
				if !config.MatchesFileName(fileName) {
					return false, false
				}
			}

			if slices.ContainsFunc(config.FileNames(), func(fn string) bool {
				return r.toPath(fn) == path
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
		tsconfig = search.Path[len(search.Path)-1].configFileName
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

	if ancestorConfigName := r.configFileRegistryBuilder.GetAncestorConfigFileName(fileName, path, configFileName, project.ProjectLoadKindCreate, nil); ancestorConfigName != "" {
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
