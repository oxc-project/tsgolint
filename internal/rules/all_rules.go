package rules

import (
	"github.com/typescript-eslint/tsgolint/internal/rule"

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
	"github.com/typescript-eslint/tsgolint/internal/rules/switch_exhaustiveness_check"
	"github.com/typescript-eslint/tsgolint/internal/rules/unbound_method"
	"github.com/typescript-eslint/tsgolint/internal/rules/use_unknown_in_catch_callback_variable"
)

var AllRules = []rule.Rule{
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
	switch_exhaustiveness_check.SwitchExhaustivenessCheckRule,
	unbound_method.UnboundMethodRule,
	use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule,
}

var AllRulesByName = make(map[string]rule.Rule, len(AllRules))

func init() {
	for _, rule := range AllRules {
		AllRulesByName[rule.Name] = rule
	}
}