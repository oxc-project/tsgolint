package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamBaseCasesBoolean tests all 26 base cases with boolean && operator
// Matches upstream: describe('base cases') > describe('and') > describe('boolean')
func TestUpstreamBaseCasesBoolean(t *testing.T) {
	// Basic && cases
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator: "&&",
	})

	// With trailing && bing
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: AddTrailingAnd,
	})...)

	// With trailing && bing.bong
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: AddTrailingAndBingBong,
	})...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		invalidCases,
	)
}

// TestUpstreamBaseCasesNotEqualNull tests != null transformations
// Matches upstream: describe('base cases') > describe('and') > describe('strict nullish equality checks') > describe('!= null')
// Note: Upstream uses suggestion fixer, but our implementation provides direct fixes
func TestUpstreamBaseCasesNotEqualNull(t *testing.T) {
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator:     "&&",
		MutateCode:   ReplaceOperatorWithNotEqualNull,
		MutateOutput: Identity, // Output doesn't change the check style
		// Note: We provide direct fixes unlike upstream which uses suggestions
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		invalidCases,
	)
}

// TestUpstreamBaseCasesNotEqualUndefined tests != undefined transformations
// Matches upstream: describe('base cases') > describe('and') > describe('strict nullish equality checks') > describe('!= undefined')
func TestUpstreamBaseCasesNotEqualUndefined(t *testing.T) {
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator:     "&&",
		MutateCode:   ReplaceOperatorWithNotEqualUndefined,
		MutateOutput: Identity,
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		invalidCases,
	)
}

// TestUpstreamBaseCasesStrictNotEqualNull tests !== null with null-only types
// Matches upstream: describe('base cases') > describe('and') > describe('strict nullish equality checks') > describe('!== null')
func TestUpstreamBaseCasesStrictNotEqualNull(t *testing.T) {
	// With | null | undefined type, !== null is NOT a valid conversion (doesn't cover undefined)
	// So these should be VALID (no error)
	validCases := GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: ReplaceOperatorWithStrictNotEqualNull,
	})

	// With | null only type (no undefined), !== null IS a valid conversion
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator:          "&&",
		MutateCode:        ReplaceOperatorWithStrictNotEqualNull,
		MutateDeclaration: RemoveUndefinedFromType,
		MutateOutput:      Identity,
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		validCases,
		invalidCases,
	)
}

// TestUpstreamBaseCasesStrictNotEqualUndefined tests !== undefined with undefined-only types
// Matches upstream: describe('base cases') > describe('and') > describe('strict nullish equality checks') > describe('!== undefined')
func TestUpstreamBaseCasesStrictNotEqualUndefined(t *testing.T) {
	// With | null | undefined type, !== undefined is NOT a valid conversion (doesn't cover null)
	// So these should be VALID (no error)
	// Skip IDs 20 and 26 as in upstream
	validCases := GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: ReplaceOperatorWithStrictNotEqualUndefined,
		SkipIDs:    map[int]bool{20: true, 26: true},
	})

	// With | undefined only type (no null), !== undefined IS a valid conversion
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator:          "&&",
		MutateCode:        ReplaceOperatorWithStrictNotEqualUndefined,
		MutateDeclaration: RemoveNullFromType,
		MutateOutput:      Identity,
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		validCases,
		invalidCases,
	)
}

// TestUpstreamBaseCasesOrBoolean tests all 26 base cases with || operator (negated)
// Matches upstream: describe('base cases') > describe('or') > describe('boolean')
func TestUpstreamBaseCasesOrBoolean(t *testing.T) {
	// OR cases need negation: !foo || !foo.bar
	// This is more complex and needs special handling
	// TODO: Implement OR boolean cases
	t.Skip("OR boolean cases need special negation handling")
}

// TestUpstreamBaseCasesOrEqualNull tests == null || transformations
// Matches upstream: describe('base cases') > describe('or') > describe('strict nullish equality checks') > describe('== null')
func TestUpstreamBaseCasesOrEqualNull(t *testing.T) {
	// foo == null || foo.bar == null -> foo?.bar == null
	// TODO: Implement OR == null cases
	t.Skip("OR == null cases need implementation")
}
