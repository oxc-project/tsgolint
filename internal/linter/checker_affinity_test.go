package linter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"gotest.tools/v3/assert"
)

func affinityTestProgram(t *testing.T, singleThreaded bool) (*compiler.Program, []*ast.SourceFile) {
	t.Helper()
	dir := filepath.ToSlash(t.TempDir())
	assert.NilError(t, os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{"compilerOptions":{"strict":true,"types":[],"target":"esnext"},"include":["*.ts"]}`), 0o600))
	for i := range 20 {
		code := fmt.Sprintf("export const value%d = { count: %d, nested: { enabled: true } };", i, i)
		assert.NilError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%02d.ts", i)), []byte(code), 0o600))
	}
	host := utils.CreateCompilerHost(dir, cachedBaseFS)
	program, diagnostics, err := utils.CreateProgram(singleThreaded, cachedBaseFS, dir, "tsconfig.json", host, false)
	assert.NilError(t, err)
	assert.Equal(t, len(diagnostics), 0)
	files := make([]*ast.SourceFile, 20)
	for i := range files {
		files[i] = program.GetSourceFile(filepath.ToSlash(filepath.Join(dir, fmt.Sprintf("file%02d.ts", i))))
		assert.Assert(t, files[i] != nil)
	}
	return program, files
}

func TestCheckerWorkloadAffinity(t *testing.T) {
	for _, single := range []bool{false, true} {
		t.Run(fmt.Sprintf("single=%t", single), func(t *testing.T) {
			program, files := affinityTestProgram(t, single)
			slices.Reverse(files)
			indices := make(map[*checker.Checker]int)
			var mu sync.Mutex
			program.ForEachCheckerParallel(func(idx int, ch *checker.Checker) {
				mu.Lock()
				defer mu.Unlock()
				indices[ch] = idx
			})
			if single {
				assert.Equal(t, len(indices), 1)
			} else {
				assert.Equal(t, len(indices), 4)
			}
			for _, selected := range [][]*ast.SourceFile{nil, files[:1], {files[1], files[6], files[11]}, files} {
				expected := make(map[*checker.Checker][]*ast.SourceFile)
				for _, file := range selected {
					ch, done := program.GetTypeCheckerForFile(t.Context(), file)
					expected[ch] = append(expected[ch], file)
					done()
				}
				lastIndex := -1
				workloads := 0
				for workload := range makeCheckerWorkloadQueue(program, selected) {
					assert.Assert(t, indices[workload.checker] > lastIndex)
					lastIndex = indices[workload.checker]
					assert.Assert(t, workload.program == program)
					var actual []*ast.SourceFile
					for file := range workload.queue {
						actual = append(actual, file)
					}
					assert.Assert(t, slices.Equal(actual, expected[workload.checker]))
					workloads++
				}
				if len(selected) == 0 {
					assert.Equal(t, workloads, 0)
				} else {
					// Empty owner queues still provide checkers that can steal.
					assert.Equal(t, workloads, len(indices))
				}
			}
		})
	}
}

type affinityUnusablePool struct{}

func (affinityUnusablePool) GetChecker(context.Context, *ast.SourceFile) (*checker.Checker, func()) {
	panic("the workload builder must not acquire request-scoped checkers")
}

func TestCheckerWorkloadCustomPool(t *testing.T) {
	dir := filepath.ToSlash(t.TempDir())
	fileName := dir + "/file.ts"
	fs := utils.NewOverlayVFS(cachedBaseFS, map[string]string{fileName: "export const value = 1;"})
	program := compiler.NewProgram(compiler.ProgramOptions{
		Config: &tsoptions.ParsedCommandLine{ParsedConfig: &core.ParsedOptions{
			CompilerOptions: &core.CompilerOptions{NoLib: core.TSTrue},
			FileNames:       []string{fileName},
		}},
		Host:              utils.CreateCompilerHost(dir, fs),
		CreateCheckerPool: func(*compiler.Program) compiler.CheckerPool { return affinityUnusablePool{} },
	})
	file := program.GetSourceFile(fileName)
	assert.Assert(t, file != nil)
	_, open := <-makeCheckerWorkloadQueue(program, []*ast.SourceFile{file})
	assert.Assert(t, !open)
}

func TestCheckerWorkloadStealing(t *testing.T) {
	files := []*ast.SourceFile{{}, {}, {}, {}}
	queues := []chan *ast.SourceFile{
		makeSourceFileQueue(files[:1]),
		makeSourceFileQueue(files[1:2]),
		makeSourceFileQueue(files[2:]),
	}
	pending := make(chan checkerWorkload, 1)
	pending <- checkerWorkload{}
	close(pending)
	w := checkerWorkload{queue: queues[0], peers: queues, pending: pending}
	assert.Assert(t, w.nextFile() == files[0])
	assert.Assert(t, w.nextFile() == nil, "unclaimed owners take precedence over stealing")
	// A peer wins the last workload between our decision to claim it and the
	// actual receive. We must retain this checker and fall back to stealing.
	<-pending
	next, ok := w.nextWorkload()
	assert.Assert(t, ok && next.queue == w.queue)
	assert.Assert(t, w.nextFile() == files[2], "steal from the largest remaining queue")
	assert.Assert(t, w.nextFile() == files[1])
	assert.Assert(t, w.nextFile() == files[3])
	assert.Assert(t, w.nextFile() == nil)
	_, ok = w.nextWorkload()
	assert.Assert(t, !ok)
}

func TestCheckerWorkloadStealingWhileOwnerBusy(t *testing.T) {
	program, files := affinityTestProgram(t, false)
	owner, done := program.GetTypeCheckerForFile(t.Context(), files[0])
	done()
	var selected []*ast.SourceFile
	for _, file := range files {
		ch, release := program.GetTypeCheckerForFile(t.Context(), file)
		if ch == owner {
			selected = append(selected, file)
		}
		release()
	}
	var busy, idle checkerWorkload
	for w := range makeCheckerWorkloadQueue(program, selected) {
		if w.checker == owner {
			busy = w
		} else {
			idle = w
		}
	}
	// The owner has started one large file. An initially empty peer can finish
	// all of its remaining files without waiting for that file to complete.
	assert.Assert(t, busy.nextFile() == selected[0])
	assert.Assert(t, idle.checker != owner)
	for _, file := range selected[1:] {
		assert.Assert(t, idle.nextFile() == file)
	}
	assert.Assert(t, idle.nextFile() == nil)
	assert.Assert(t, busy.nextFile() == nil)
}

func TestRunLinterOnProgram_StealFromSingleOwner(t *testing.T) {
	for _, timed := range []bool{false, true} {
		t.Run(fmt.Sprintf("timed=%t", timed), func(t *testing.T) {
			program, files := affinityTestProgram(t, false)
			owner, release := program.GetTypeCheckerForFile(t.Context(), files[0])
			release()
			var selected []*ast.SourceFile
			for _, file := range files {
				ch, done := program.GetTypeCheckerForFile(t.Context(), file)
				if ch == owner {
					selected = append(selected, file)
				}
				done()
			}
			assert.Assert(t, len(selected) >= 4)
			var started atomic.Int32
			gate := make(chan struct{})
			used := make(map[*checker.Checker]bool)
			visits := make(map[*ast.SourceFile]int)
			var mu sync.Mutex
			var timings *RuleTimingStore
			if timed {
				timings = NewRuleTimingStore()
			}
			err := RunLinterOnProgram(RunLinterOnProgramOptions{
				Program: program, Files: selected, Workers: 4, TimingStore: timings,
				GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
					return []ConfiguredRule{{Name: "stealing", Run: func(ctx rule.RuleContext) rule.RuleListeners {
						mu.Lock()
						used[ctx.TypeChecker] = true
						visits[ctx.SourceFile]++
						mu.Unlock()
						if n := started.Add(1); n <= 4 {
							if n == 4 {
								close(gate)
							}
							select {
							case <-gate:
							case <-time.After(5 * time.Second):
								t.Error("idle checkers did not steal while the owner was busy")
							}
						}
						return nil
					}}}
				},
				OnDiagnostic:         func(rule.RuleDiagnostic) {},
				OnInternalDiagnostic: func(d diagnostic.Internal) { t.Errorf("unexpected diagnostic: %s", d.Description) },
			})
			assert.NilError(t, err)
			assert.Equal(t, len(used), 4)
			assert.Equal(t, len(visits), len(selected))
			for _, file := range selected {
				assert.Equal(t, visits[file], 1)
			}
		})
	}
}

func TestRunLinterOnProgram_SingleFileKeepsOwner(t *testing.T) {
	program, files := affinityTestProgram(t, false)
	for idx, file := range files {
		for _, timed := range []bool{false, true} {
			t.Run(fmt.Sprintf("file=%d/timed=%t", idx, timed), func(t *testing.T) {
				owner, release := program.GetTypeCheckerForFile(t.Context(), file)
				release()
				var timings *RuleTimingStore
				if timed {
					timings = NewRuleTimingStore()
				}
				visits := 0
				err := RunLinterOnProgram(RunLinterOnProgramOptions{
					Program: program, Files: []*ast.SourceFile{file}, Workers: 8, TimingStore: timings,
					GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
						return []ConfiguredRule{{Name: "single", Run: func(ctx rule.RuleContext) rule.RuleListeners {
							if ctx.TypeChecker != owner {
								t.Error("a single file was stolen from its warmed checker")
							}
							visits++
							return nil
						}}}
					},
					OnDiagnostic:         func(rule.RuleDiagnostic) {},
					OnInternalDiagnostic: func(d diagnostic.Internal) { t.Errorf("unexpected diagnostic: %s", d.Description) },
				})
				assert.NilError(t, err)
				assert.Equal(t, visits, 1)
			})
		}
	}
}

func TestRunLinterOnProgram_CheckerAffinity(t *testing.T) {
	for _, workers := range []int{1, 2, 4, 8} {
		for _, semantic := range []bool{false, true} {
			for _, timed := range []bool{false, true} {
				t.Run(fmt.Sprintf("workers=%d/semantic=%t/timed=%t", workers, semantic, timed), func(t *testing.T) {
					program, files := affinityTestProgram(t, false)
					files = files[1:18:18]
					expected := make(map[*ast.SourceFile]*checker.Checker)
					busy := make(map[*checker.Checker]*atomic.Int32)
					for _, file := range files {
						ch, done := program.GetTypeCheckerForFile(t.Context(), file)
						expected[file] = ch
						busy[ch] = &atomic.Int32{}
						done()
					}
					var active atomic.Int32
					visits := make(map[*ast.SourceFile]int)
					var mu sync.Mutex
					var timings *RuleTimingStore
					if timed {
						timings = NewRuleTimingStore()
					}
					err := RunLinterOnProgram(RunLinterOnProgramOptions{
						Program: program, Files: files, Workers: workers, TimingStore: timings,
						TypeErrors: TypeErrors{ReportSemantic: semantic},
						GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
							return []ConfiguredRule{{Name: "affinity", Run: func(ctx rule.RuleContext) rule.RuleListeners {
								if workers == 1 && ctx.TypeChecker != expected[ctx.SourceFile] {
									t.Error("file was assigned to a different checker")
									return nil
								}
								return rule.RuleListeners{ast.KindVariableDeclaration: func(node *ast.Node) {
									if busy[ctx.TypeChecker].Add(1) != 1 {
										t.Error("concurrent use of the same checker")
									}
									defer busy[ctx.TypeChecker].Add(-1)
									if active.Add(1) > int32(workers) {
										t.Error("worker limit exceeded")
									}
									defer active.Add(-1)
									ctx.TypeChecker.GetTypeAtLocation(node)
									mu.Lock()
									visits[ctx.SourceFile]++
									mu.Unlock()
								}}
							}}}
						},
						OnDiagnostic:         func(rule.RuleDiagnostic) {},
						OnInternalDiagnostic: func(d diagnostic.Internal) { t.Errorf("unexpected diagnostic: %s", d.Description) },
					})
					assert.NilError(t, err)
					assert.Equal(t, len(visits), len(files))
					for _, file := range files {
						assert.Equal(t, visits[file], 1)
					}
					if timed {
						records := timings.Collect()
						assert.Equal(t, len(records), 1)
						assert.Equal(t, records[0].Calls, uint64(len(files)*2))
					}
				})
			}
		}
	}
}
