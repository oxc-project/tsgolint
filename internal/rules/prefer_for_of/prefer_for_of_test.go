package prefer_for_of

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferForOfRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferForOfRule, []rule_tester.ValidTestCase{
		// Valid cases - should NOT trigger the rule
		{Code: `
// for-of loop - already good
for (const item of array) {
  console.log(item);
}
    `},
		{Code: `
// Loop variable used to index other arrays - should not convert
type Series = { data: Array<{ x: number; y: number }> };

function a(series: Series, substract: Series[]): void {
  for (let x = 0; x < series.data.length; x++) {
    let newValue = series.data[x]!.y;
    for (const otherSeries of substract) {
      newValue -= otherSeries.data[x]!.y; // x is used to index other arrays
    }
    series.data[x]!.y = newValue;
  }
}
    `},
		{Code: `
// Loop variable used in other contexts
const array = [1, 2, 3];
for (let i = 0; i < array.length; i++) {
  console.log(i); // Using i directly, not array[i]
}
    `},
		{Code: `
// Non-standard for loop pattern
const array = [1, 2, 3];
for (let i = 1; i < array.length; i++) { // starts at 1, not 0
  console.log(array[i]);
}
    `},
		{Code: `
// Non-standard increment
const array = [1, 2, 3];
for (let i = 0; i < array.length; i += 2) { // increment by 2
  console.log(array[i]);
}
    `},
		{Code: `
// Loop without array access
const array = [1, 2, 3];
for (let i = 0; i < array.length; i++) {
  doSomething(); // not using array[i]
}
    `},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases - should trigger the rule
		{
			Code: `
const array = [1, 2, 3];
for (let i = 0; i < array.length; i++) {
  console.log(array[i]);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferForOf",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 39,
				},
			},
		},
		{
			Code: `
function processArray(items: number[]) {
  for (let i = 0; i < items.length; i++) {
    const value = items[i];
    console.log(value);
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferForOf",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 41,
				},
			},
		},
	})
}