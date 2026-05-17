package linter

import (
	"log"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func RunLinterOnProgramWithTimings(logLevel utils.LogLevel, program *compiler.Program, files []*ast.SourceFile, workers int, getRulesForFile func(sourceFile *ast.SourceFile) []ConfiguredRule, onDiagnostic func(diagnostic rule.RuleDiagnostic), onInternalDiagnostic func(d diagnostic.Internal), fixState Fixes, typeErrors TypeErrors, timingStore *RuleTimingStore) error {
	reportTypeScriptDiagnostics(program, files, typeErrors, onInternalDiagnostic)
	workloadQueue := makeCheckerWorkloadQueue(program, files)

	wg := core.NewWorkGroup(workers == 1)
	for range workers {
		wg.Queue(func() {
			type taggedListener struct {
				ruleName string
				ruleIdx  int
				fn       func(node *ast.Node)
			}
			registeredListeners := make(map[ast.Kind][]taggedListener, 20)
			localTimings := make(map[string]RuleTimingStat, 64)

			recordTiming := func(stat *RuleTimingStat, duration time.Duration) {
				stat.Duration += duration
				stat.Calls++
			}

			ctxBuilder := &ruleContextBuilder{
				fixState:     fixState,
				onDiagnostic: onDiagnostic,
			}

			// These closures remain valid for the length of linting, as we mutate the fields
			// of `ctxBuilder`, but `ctxBuilder` itself will not change.
			ctx := newRuleContext(ctxBuilder)

			for w := range workloadQueue {
				ctxBuilder.program = w.program
				ctxBuilder.checker = w.checker
				ctx.Program = w.program
				ctx.TypeChecker = w.checker

				for file := range w.queue {
					if logLevel == utils.LogLevelDebug {
						log.Print(file.FileName())
					}
					ctxBuilder.file = file
					ctx.SourceFile = file

					rules := getRulesForFile(file)
					timingStats := make([]RuleTimingStat, len(rules))
					for ruleIdx, r := range rules {
						ctxBuilder.ruleName = r.Name
						start := time.Now()
						listenersByKind := r.Run(ctx)
						recordTiming(&timingStats[ruleIdx], time.Since(start))
						for kind, listener := range listenersByKind {
							listeners, ok := registeredListeners[kind]
							if !ok {
								listeners = make([]taggedListener, 0, len(rules))
							}
							registeredListeners[kind] = append(listeners, taggedListener{ruleName: r.Name, ruleIdx: ruleIdx, fn: listener})
						}
					}

					runListeners := func(kind ast.Kind, node *ast.Node) {
						if listeners, ok := registeredListeners[kind]; ok {
							for _, listener := range listeners {
								ctxBuilder.ruleName = listener.ruleName
								start := time.Now()
								listener.fn(node)
								recordTiming(&timingStats[listener.ruleIdx], time.Since(start))
							}
						}
					}

					visitLintNodes(file, runListeners)
					for idx, stat := range timingStats {
						if stat.Calls == 0 {
							continue
						}
						merged := localTimings[rules[idx].Name]
						merged.add(stat)
						localTimings[rules[idx].Name] = merged
					}
					// Instead of clearing the map, we clear the slices in-place to avoid re-allocating memory for the listeners on each file.
					for k := range registeredListeners {
						registeredListeners[k] = registeredListeners[k][:0]
					}
				}
			}

			timingStore.merge(localTimings)
		})
	}
	wg.RunAndWait()

	return nil
}
