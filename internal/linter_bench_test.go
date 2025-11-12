// Run benchmark of the linter against a specified directory.
// Example: `TSGOLINT_BENCH_DIR=/some/directory go test -bench=BenchmarkLinter -benchtime=1x ./internal/`
package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"github.com/typescript-eslint/tsgolint/internal/rules/await_thenable"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_array_delete"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_base_to_string"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_confusing_void_expression"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_deprecated"
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
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

var allRules = []rule.Rule{
	await_thenable.AwaitThenableRule,
	no_array_delete.NoArrayDeleteRule,
	no_base_to_string.NoBaseToStringRule,
	no_confusing_void_expression.NoConfusingVoidExpressionRule,
	no_deprecated.NoDeprecatedRule,
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

// BenchmarkLinter benchmarks the linter against a specified directory.
//
// To run this benchmark, set the TSGOLINT_BENCH_DIR environment variable
// to the directory you want to benchmark against. The directory should contain
// a TypeScript project with a tsconfig.json file.
//
// Example:
//
//	TSGOLINT_BENCH_DIR=/path/to/project go test -bench=BenchmarkLinter -benchtime=3x ./internal/
//
// By default (if TSGOLINT_BENCH_DIR is not set), it will benchmark against
// the e2e/fixtures/basic directory.
func BenchmarkLinter(b *testing.B) {
	benchDir := os.Getenv("TSGOLINT_BENCH_DIR")
	if benchDir == "" {
		// Default to e2e/fixtures/basic
		wd, err := os.Getwd()
		if err != nil {
			b.Fatalf("failed to get working directory: %v", err)
		}
		// Navigate up to the project root and find e2e/fixtures/basic
		benchDir = filepath.Join(wd, "..", "e2e", "fixtures", "basic")
	}

	// Normalize the path
	benchDir, err := filepath.Abs(benchDir)
	if err != nil {
		b.Fatalf("failed to get absolute path: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(benchDir); os.IsNotExist(err) {
		b.Skipf("benchmark directory does not exist: %s. Set TSGOLINT_BENCH_DIR to a valid directory", benchDir)
	}

	benchDir = tspath.NormalizePath(benchDir)

	// Setup filesystem and config
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	configFileName := tspath.ResolvePath(benchDir, "tsconfig.json")

	if !fs.FileExists(configFileName) {
		// Create a default tsconfig.json overlay if it doesn't exist
		fs = utils.NewOverlayVFS(fs, map[string]string{
			configFileName: "{}",
		})
	}

	currentDirectory := tspath.GetDirectoryPath(configFileName)
	host := utils.CreateCompilerHost(currentDirectory, fs)

	// Create program
	program, _, err := utils.CreateProgram(false, fs, currentDirectory, configFileName, host)
	if err != nil {
		b.Fatalf("error creating TS program: %v", err)
	}

	if program == nil {
		b.Fatal("error creating TS program: program is nil")
	}

	// Filter files (exclude node_modules)
	var files []*ast.SourceFile
	cwdPath := string(tspath.ToPath("", currentDirectory, program.Host().FS().UseCaseSensitiveFileNames()).EnsureTrailingDirectorySeparator())
	for _, file := range program.SourceFiles() {
		p := string(file.Path())
		if strings.Contains(p, "/node_modules/") {
			continue
		}
		if _, matched := strings.CutPrefix(p, cwdPath); matched {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Skipf("no files to benchmark in directory: %s", benchDir)
	}

	b.Logf("Benchmarking %d files with %d rules", len(files), len(allRules))
	b.ResetTimer()

	for i := range b.N {
		diagnosticCount := 0

		err := linter.RunLinterOnProgram(
			utils.LogLevelNormal,
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
				diagnosticCount++
			},
			linter.Fixes{
				Fix:            true,
				FixSuggestions: true,
			},
		)

		if err != nil {
			b.Fatalf("error running linter: %v", err)
		}

		if i == 0 {
			b.Logf("Found %d diagnostics", diagnosticCount)
		}
	}
}
