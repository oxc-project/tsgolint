package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"github.com/typescript-eslint/tsgolint/internal/rules/await_thenable"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_array_delete"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_base_to_string"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_confusing_void_expression"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_duplicate_type_constituents"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_floating_promises"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_for_in_array"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_implied_eval"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_meaningless_void_operator"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_misused_promises"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_misused_spread"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_mixed_enums"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_redundant_type_constituents"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_boolean_literal_compare"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_template_expression"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_type_arguments"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_type_assertion"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_argument"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_assignment"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_call"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_enum_comparison"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_member_access"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_return"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_type_assertion"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_unary_minus"
	"github.com/typescript-eslint/tsgolint/internal/rules/non_nullable_type_assertion_style"
	"github.com/typescript-eslint/tsgolint/internal/rules/only_throw_error"
	"github.com/typescript-eslint/tsgolint/internal/rules/prefer_promise_reject_errors"
	"github.com/typescript-eslint/tsgolint/internal/rules/prefer_reduce_type_parameter"
	"github.com/typescript-eslint/tsgolint/internal/rules/prefer_return_this_type"
	"github.com/typescript-eslint/tsgolint/internal/rules/promise_function_async"
	"github.com/typescript-eslint/tsgolint/internal/rules/related_getter_setter_pairs"
	"github.com/typescript-eslint/tsgolint/internal/rules/require_array_sort_compare"
	"github.com/typescript-eslint/tsgolint/internal/rules/require_await"
	"github.com/typescript-eslint/tsgolint/internal/rules/restrict_plus_operands"
	"github.com/typescript-eslint/tsgolint/internal/rules/restrict_template_expressions"
	"github.com/typescript-eslint/tsgolint/internal/rules/return_await"
	"github.com/typescript-eslint/tsgolint/internal/rules/strict_boolean_expressions"
	"github.com/typescript-eslint/tsgolint/internal/rules/switch_exhaustiveness_check"
	"github.com/typescript-eslint/tsgolint/internal/rules/unbound_method"
	"github.com/typescript-eslint/tsgolint/internal/rules/use_unknown_in_catch_callback_variable"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

func recordTrace(traceOut string) (func(), error) {
	if traceOut != "" {
		f, err := os.Create(traceOut)
		if err != nil {
			return nil, fmt.Errorf("error creating trace file: %w", err)
		}
		trace.Start(f)
		return func() {
			trace.Stop()
			f.Close()
		}, nil
	}
	return func() {}, nil
}
func recordCpuprof(cpuprofOut string) (func(), error) {
	if cpuprofOut != "" {
		f, err := os.Create(cpuprofOut)
		if err != nil {
			return nil, fmt.Errorf("error creating cpuprof file: %w", err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return nil, fmt.Errorf("error starting cpu profiling: %w", err)
		}
		return func() {
			pprof.StopCPUProfile()
			f.Close()
		}, nil
	}
	return func() {}, nil
}

func writeMemProfiles(heapOut string, allocsOut string) {
	if heapOut != "" {
		if f, err := os.Create(heapOut); err == nil {
			_ = pprof.WriteHeapProfile(f)
			_ = f.Close()
		}
	}

	if allocsOut != "" {
		if f, err := os.Create(allocsOut); err == nil {
			// debug=0 → compressed protobuf suitable for pprof
			_ = pprof.Lookup("allocs").WriteTo(f, 0)
			_ = f.Close()
		}
	}
}

var allRules = []rule.Rule{
	await_thenable.AwaitThenableRule,
	no_array_delete.NoArrayDeleteRule,
	no_base_to_string.NoBaseToStringRule,
	no_confusing_void_expression.NoConfusingVoidExpressionRule,
	no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule,
	no_floating_promises.NoFloatingPromisesRule,
	no_for_in_array.NoForInArrayRule,
	no_implied_eval.NoImpliedEvalRule,
	no_meaningless_void_operator.NoMeaninglessVoidOperatorRule,
	no_misused_promises.NoMisusedPromisesRule,
	no_misused_spread.NoMisusedSpreadRule,
	no_mixed_enums.NoMixedEnumsRule,
	no_redundant_type_constituents.NoRedundantTypeConstituentsRule,
	no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule,
	no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule,
	no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule,
	no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule,
	no_unsafe_argument.NoUnsafeArgumentRule,
	no_unsafe_assignment.NoUnsafeAssignmentRule,
	no_unsafe_call.NoUnsafeCallRule,
	no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule,
	no_unsafe_member_access.NoUnsafeMemberAccessRule,
	no_unsafe_return.NoUnsafeReturnRule,
	no_unsafe_type_assertion.NoUnsafeTypeAssertionRule,
	no_unsafe_unary_minus.NoUnsafeUnaryMinusRule,
	non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule,
	only_throw_error.OnlyThrowErrorRule,
	prefer_promise_reject_errors.PreferPromiseRejectErrorsRule,
	prefer_reduce_type_parameter.PreferReduceTypeParameterRule,
	prefer_return_this_type.PreferReturnThisTypeRule,
	promise_function_async.PromiseFunctionAsyncRule,
	related_getter_setter_pairs.RelatedGetterSetterPairsRule,
	require_array_sort_compare.RequireArraySortCompareRule,
	require_await.RequireAwaitRule,
	restrict_plus_operands.RestrictPlusOperandsRule,
	restrict_template_expressions.RestrictTemplateExpressionsRule,
	return_await.ReturnAwaitRule,
	strict_boolean_expressions.StrictBooleanExpressionsRule,
	switch_exhaustiveness_check.SwitchExhaustivenessCheckRule,
	unbound_method.UnboundMethodRule,
	use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule,
}

var allRulesByName = make(map[string]rule.Rule, len(allRules))

func init() {
	for _, rule := range allRules {
		allRulesByName[rule.Name] = rule
	}
}

const spaces = "                                                                                                    "

func printDiagnostic(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions) {
	diagnosticStart := d.Range.Pos()
	diagnosticEnd := d.Range.End()

	diagnosticStartLine, diagnosticStartColumn := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
	diagnosticEndline, _ := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

	lineMap := d.SourceFile.ECMALineMap()
	text := d.SourceFile.Text()

	codeboxStartLine := max(diagnosticStartLine-1, 0)
	codeboxEndLine := min(diagnosticEndline+1, len(lineMap)-1)

	codeboxStart := scanner.GetECMAPositionOfLineAndCharacter(d.SourceFile, codeboxStartLine, 0)
	var codeboxEndColumn int
	if codeboxEndLine == len(lineMap)-1 {
		codeboxEndColumn = len(text) - int(lineMap[len(lineMap)-1])
	} else {
		codeboxEndColumn = int(lineMap[codeboxEndLine+1]-lineMap[codeboxEndLine]) - 1
	}
	codeboxEnd := scanner.GetECMAPositionOfLineAndCharacter(d.SourceFile, codeboxEndLine, codeboxEndColumn)

	w.Write([]byte{' ', 0x1b, '[', '7', 'm', 0x1b, '[', '1', 'm', 0x1b, '[', '3', '8', ';', '5', ';', '3', '7', 'm', ' '})
	w.WriteString(d.RuleName)
	w.WriteString(" \x1b[0m — ")
	messageLineStart := 0
	for i, char := range d.Message.Description {
		if char == '\n' {
			w.WriteString(d.Message.Description[messageLineStart : i+1])
			messageLineStart = i + 1
			w.WriteString("    \x1b[2m│\x1b[0m")
			w.WriteString(spaces[:len(d.RuleName)+1])
		}
	}
	if messageLineStart <= len(d.Message.Description) {
		w.WriteString(d.Message.Description[messageLineStart:len(d.Message.Description)])
	}
	w.WriteString("\n  \x1b[2m╭─┴──────────(\x1b[0m \x1b[3m\x1b[38;5;117m")
	w.WriteString(tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions))
	w.WriteByte(':')
	w.WriteString(strconv.Itoa(diagnosticStartLine + 1))
	w.WriteByte(':')
	w.WriteString(strconv.Itoa(diagnosticStartColumn + 1))
	w.WriteString("\x1b[0m \x1b[2m)─────\x1b[0m\n")

	indentSize := math.MaxInt
	line := codeboxStartLine
	lineIndentCalculated := false
	lastNonSpaceIndex := -1

	lineStarts := make([]int, 13)
	lineEnds := make([]int, 13)

	if codeboxEndLine-codeboxStartLine >= len(lineEnds) {
		w.WriteString("  \x1b[2m│\x1b[0m  Error range is too big. Skipping code block printing.\n  \x1b[2m╰────────────────────────────────\x1b[0m\n\n")
		return
	}

	for i, char := range text[codeboxStart:codeboxEnd] {
		if char == '\n' {
			if line != codeboxEndLine {
				lineIndentCalculated = false
				lineEnds[line-codeboxStartLine] = lastNonSpaceIndex - int(lineMap[line]) + codeboxStart
				lastNonSpaceIndex = -1
				line++
			}
			continue
		}

		if !lineIndentCalculated && !unicode.IsSpace(char) {
			lineIndentCalculated = true
			lineStarts[line-codeboxStartLine] = i - int(lineMap[line]) + codeboxStart
			indentSize = min(indentSize, lineStarts[line-codeboxStartLine])
		}

		if lineIndentCalculated && !unicode.IsSpace(char) {
			lastNonSpaceIndex = i + 1
		}
	}
	if line == codeboxEndLine {
		lineEnds[line-codeboxStartLine] = lastNonSpaceIndex - int(lineMap[line]) + codeboxStart
	}

	diagnosticHighlightActive := false
	lastLineNumber := strconv.Itoa(codeboxEndLine + 1)
	for line := codeboxStartLine; line <= codeboxEndLine; line++ {
		w.WriteString("  \x1b[2m│ ")
		if line == codeboxEndLine {
			w.WriteString(lastLineNumber)
		} else {
			number := strconv.Itoa(line + 1)
			if len(number) < len(lastLineNumber) {
				w.WriteByte(' ')
			}
			w.WriteString(number)
		}
		w.WriteString(" │\x1b[0m  ")

		lineTextStart := int(lineMap[line]) + indentSize
		underlineStart := max(lineTextStart, int(lineMap[line])+lineStarts[line-codeboxStartLine])
		underlineEnd := underlineStart
		lineTextEnd := max(int(lineMap[line])+lineEnds[line-codeboxStartLine], lineTextStart)

		if diagnosticHighlightActive {
			underlineEnd = lineTextEnd
		} else if int(lineMap[line]) <= diagnosticStart && (line == len(lineMap) || diagnosticStart < int(lineMap[line+1])) {
			underlineStart = min(max(lineTextStart, diagnosticStart), lineTextEnd)
			underlineEnd = lineTextEnd
			diagnosticHighlightActive = true
		}
		if int(lineMap[line]) <= diagnosticEnd && (line == len(lineMap) || diagnosticEnd < int(lineMap[line+1])) {
			underlineEnd = min(max(underlineStart, diagnosticEnd), lineTextEnd)
			diagnosticHighlightActive = false
		}

		if underlineStart != underlineEnd {
			w.WriteString(text[lineTextStart:underlineStart])
			w.Write([]byte{
				0x1b, '[', '4', 'm',
				0x1b, '[', '4', ':', '3', 'm',
				0x1b, '[', '5', '8', ':', '5', ':', '1', '6', '0', 'm',
				0x1b, '[', '3', '8', ';', '5', ';', '1', '6', '0', 'm',
				0x1b, '[', '2', '2', ';', '4', '9', 'm',
			})
			w.WriteString(text[underlineStart:underlineEnd])
			w.Write([]byte{0x1b, '[', '0', 'm'})
			w.WriteString(text[underlineEnd:lineTextEnd])
		} else if lineTextStart != lineTextEnd {
			w.WriteString(text[lineTextStart:lineTextEnd])
		}

		w.WriteByte('\n')
	}
	w.WriteString("  \x1b[2m╰────────────────────────────────\x1b[0m\n\n")
}

const usage = `✨ tsgolint - speedy TypeScript linter

Usage:
    tsgolint [OPTIONS]

Options:
    --tsconfig PATH   Which tsconfig to use. Defaults to tsconfig.json.
		--list-files      List matched files
    -h, --help        Show help
`

func runMain() int {
	if len(os.Args) > 1 && os.Args[1] == "headless" {
		return runHeadless(os.Args[2:])
	}

	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var (
		help      bool
		tsconfig  string
		listFiles bool

		traceOut       string
		cpuprofOut     string
		singleThreaded bool
	)

	flag.StringVar(&tsconfig, "tsconfig", "", "which tsconfig to use")
	flag.BoolVar(&listFiles, "list-files", false, "list matched files")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&help, "h", false, "show help")

	flag.StringVar(&traceOut, "trace", "", "file to put trace to")
	flag.StringVar(&cpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	flag.BoolVar(&singleThreaded, "singleThreaded", false, "run in single threaded mode")

	flag.Parse()

	if help {
		flag.Usage()
		return 0
	}

	enableVirtualTerminalProcessing()
	timeBefore := time.Now()

	if done, err := recordTrace(traceOut); err != nil {
		os.Stderr.WriteString(err.Error())
		return 1
	} else {
		defer done()
	}
	if done, err := recordCpuprof(cpuprofOut); err != nil {
		os.Stderr.WriteString(err.Error())
		return 1
	} else {
		defer done()
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
		return 1
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	var configFileName string
	if tsconfig == "" {
		configFileName = tspath.ResolvePath(currentDirectory, "tsconfig.json")
		if !fs.FileExists(configFileName) {
			fs = utils.NewOverlayVFS(fs, map[string]string{
				configFileName: "{}",
			})
		}
	} else {
		configFileName = tspath.ResolvePath(currentDirectory, tsconfig)
		if !fs.FileExists(configFileName) {
			fmt.Fprintf(os.Stderr, "error: tsconfig %q doesn't exist", tsconfig)
			return 1
		}
	}

	currentDirectory = tspath.GetDirectoryPath(configFileName)

	host := utils.CreateCompilerHost(currentDirectory, fs)

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}

	program, _, err := utils.CreateProgram(singleThreaded, fs, currentDirectory, configFileName, host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating TS program: %v", err)
		return 1
	}

	files := []*ast.SourceFile{}
	cwdPath := string(tspath.ToPath("", currentDirectory, program.Host().FS().UseCaseSensitiveFileNames()).EnsureTrailingDirectorySeparator())
	var matchedFiles strings.Builder
	for _, file := range program.SourceFiles() {
		p := string(file.Path())
		if strings.Contains(p, "/node_modules/") {
			continue
		}
		if fileName, matched := strings.CutPrefix(p, cwdPath); matched {
			if listFiles {
				matchedFiles.WriteString("Found file: ")
				matchedFiles.WriteString(fileName)
				matchedFiles.WriteByte('\n')
			}
			files = append(files, file)
		}
	}
	if listFiles {
		os.Stdout.WriteString(matchedFiles.String())
	}
	slices.SortFunc(files, func(a *ast.SourceFile, b *ast.SourceFile) int {
		return len(b.Text()) - len(a.Text())
	})

	var wg sync.WaitGroup

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
	errorsCount := 0

	wg.Go(func() {
		w := bufio.NewWriterSize(os.Stdout, 4096*100)
		defer w.Flush()
		for d := range diagnosticsChan {
			errorsCount++
			if errorsCount == 1 {
				w.WriteByte('\n')
			}
			printDiagnostic(d, w, comparePathOptions)
			if w.Available() < 4096 {
				w.Flush()
			}
		}
	})

	err = linter.RunLinterOnProgram(
		utils.GetLogLevel(),
		program,
		files,
		runtime.GOMAXPROCS(0),
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return utils.Map(allRules, func(r rule.Rule) linter.ConfiguredRule {
				return linter.ConfiguredRule{
					Name: r.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, nil)
					},
				}
			})
		},
		func(d rule.RuleDiagnostic) {
			diagnosticsChan <- d
		},
		linter.Fixes{
			Fix:            true,
			FixSuggestions: true,
		},
	)

	close(diagnosticsChan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running linter: %v\n", err)
		return 1
	}

	wg.Wait()

	errorsColor := "\x1b[1m"
	if errorsCount == 0 {
		errorsColor = "\x1b[1;32m"
	}
	errorsText := "errors"
	if errorsCount == 1 {
		errorsText = "error"
	}
	filesText := "files"
	if len(files) == 1 {
		filesText = "file"
	}
	rulesText := "rules"
	if len(allRules) == 1 {
		rulesText = "rule"
	}
	threadsCount := 1
	if !singleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}
	fmt.Fprintf(
		os.Stdout,
		"Found %v%v\x1b[0m %v \x1b[2m(linted \x1b[1m%v\x1b[22m\x1b[2m %v with \x1b[1m%v\x1b[22m\x1b[2m %v in \x1b[1m%v\x1b[22m\x1b[2m using \x1b[1m%v\x1b[22m\x1b[2m threads)\n",
		errorsColor,
		errorsCount,
		errorsText,
		len(files),
		filesText,
		len(allRules),
		rulesText,
		time.Since(timeBefore).Round(time.Millisecond),
		threadsCount,
	)

	return 0
}

func main() {
	os.Exit(runMain())
}
