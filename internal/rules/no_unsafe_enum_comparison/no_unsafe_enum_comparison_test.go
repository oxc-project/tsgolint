package no_unsafe_enum_comparison

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnsafeEnumComparisonRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeEnumComparisonRule, []rule_tester.ValidTestCase{
		{Code: "'a' > 'b';"},
		{Code: "'a' < 'b';"},
		{Code: "'a' == 'b';"},
		{Code: "'a' === 'b';"},
		{Code: "1 > 2;"},
		{Code: "1 < 2;"},
		{Code: "1 == 2;"},
		{Code: "1 === 2;"},
		{Code: `
      enum Fruit {
        Apple,
      }
      Fruit.Apple === ({} as any);
    `},
		{Code: `
      enum Fruit {
        Apple,
      }
      Fruit.Apple === undefined;
    `},
		{Code: `
      enum Fruit {
        Apple,
      }
      Fruit.Apple === null;
    `},
		{Code: `
      enum Fruit {
        Apple,
      }
      declare const fruit: Fruit | -1;
      fruit === -1;
    `},
		{Code: `
      enum Fruit {
        Apple,
      }
      declare const fruit: Fruit | number;
      fruit === -1;
    `},
		{Code: `
      enum Fruit {
        Apple,
      }
      declare const fruit: Fruit | 'apple';
      fruit === 'apple';
    `},
		{Code: `
      enum Fruit {
        Apple,
      }
      declare const fruit: Fruit | string;
      fruit === 'apple';
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
      }
      declare const fruit: Fruit | 'apple';
      fruit === 'apple';
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
      }
      declare const fruit: Fruit | string;
      fruit === 'apple';
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
      }
      declare const fruit: Fruit | 0;
      fruit === 0;
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
      }
      declare const fruit: Fruit | number;
      fruit === 0;
    `},
		{Code: `
      enum Fruit {
        Apple,
      }
      declare const fruit: Fruit | 'apple';
      fruit === Math.random() > 0.5 ? 'apple' : Fruit.Apple;
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
      }
      declare const fruit: Fruit | 'apple';
      fruit === Math.random() > 0.5 ? 'apple' : Fruit.Apple;
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
      }
      declare const fruit: Fruit | string;
      fruit === Math.random() > 0.5 ? 'apple' : Fruit.Apple;
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
      }
      declare const fruit: Fruit | 0;
      fruit === Math.random() > 0.5 ? 0 : Fruit.Apple;
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
      }
      declare const fruit: Fruit | number;
      fruit === Math.random() > 0.5 ? 0 : Fruit.Apple;
    `},
		{Code: `
      enum Fruit {
        Apple,
        Banana,
      }
      Fruit.Apple === Fruit.Banana;
    `},
		{Code: `
      enum Fruit {
        Apple = 0,
        Banana = 1,
      }
      Fruit.Apple === Fruit.Banana;
    `},
		{Code: `
      enum Fruit {
        Apple = 'apple',
        Banana = 'banana',
      }
      Fruit.Apple === Fruit.Banana;
    `},
		{Code: `
      enum Fruit {
        Apple,
        Banana,
      }
      const fruit = Fruit.Apple;
      fruit === Fruit.Banana;
    `},
		{Code: `
      enum Vegetable {
        Asparagus = 'asparagus',
        Beet = 'beet',
        Celery = 'celery',
      }
      const vegetable = Vegetable.Asparagus;
      vegetable === Vegetable.Beet;
    `},
		{Code: `
      enum Fruit {
        Apple,
        Banana,
        Cherry,
      }
      const fruit1 = Fruit.Apple;
      const fruit2 = Fruit.Banana;
      fruit1 === fruit2;
    `},
		{Code: `
      enum Vegetable {
        Asparagus = 'asparagus',
        Beet = 'beet',
        Celery = 'celery',
      }
      const vegetable1 = Vegetable.Asparagus;
      const vegetable2 = Vegetable.Beet;
      vegetable1 === vegetable2;
    `},
		{Code: `
      enum Fruit {
        Apple,
        Banana,
        Cherry,
      }
      enum Fruit2 {
        Apple2,
        Banana2,
        Cherry2,
      }
      declare const left: number | Fruit;
      declare const right: number | Fruit2;
      left === right;
    `},
		{Code: `
      enum Vegetable {
        Asparagus = 'asparagus',
        Beet = 'beet',
        Celery = 'celery',
      }
      enum Vegetable2 {
        Asparagus2 = 'asparagus2',
        Beet2 = 'beet2',
        Celery2 = 'celery2',
      }
      declare const left: string | Vegetable;
      declare const right: string | Vegetable2;
      left === right;
    `},
		{Code: `
      enum Vegetable {
        Asparagus = 'asparagus',
        Beet = 'beet',
        Celery = 'celery',
      }
      const foo = {};
      const vegetable = Vegetable.Asparagus;
      vegetable in foo;
    `},
		{Code: `
      enum Fruit {
        Apple,
        Banana,
        Cherry,
      }
      declare const fruitOrBoolean: Fruit | boolean;
      fruitOrBoolean === true;
    `},
		{Code: `
      enum Str {
        A = 'a',
      }
      enum Num {
        B = 1,
      }
      enum Mixed {
        A = 'a',
        B = 1,
      }

      declare const str: Str;
      declare const strOrString: Str | string;

      declare const num: Num;
      declare const numOrNumber: Num | number;

      declare const mixed: Mixed;
      declare const mixedOrStringOrNumber: Mixed | string | number;

      function someFunction() {}

      // following are all ignored due to the presence of "| string" or "| number"
      strOrString === 'a';
      numOrNumber === 1;
      mixedOrStringOrNumber === 'a';
      mixedOrStringOrNumber === 1;

      // following are all ignored because the value can never be an enum value
      str === 1;
      num === 'a';
      str === {};
      num === {};
      mixed === {};
      str === true;
      num === true;
      mixed === true;
      str === someFunction;
      num === someFunction;
      mixed === someFunction;
    `},
		{Code: `
      enum Fruit {
        Apple,
      }

      const bitShift = 1 << Fruit.Apple;
    `},
		{Code: `
      enum Fruit {
        Apple,
      }

      const bitShift = 1 >> Fruit.Apple;
    `},
		{Code: `
      enum Fruit {
        Apple,
      }

      declare const fruit: Fruit;

      switch (fruit) {
        case Fruit.Apple: {
          break;
        }
      }
    `},
		{Code: `
      enum Vegetable {
        Asparagus = 'asparagus',
      }

      declare const vegetable: Vegetable;

      switch (vegetable) {
        case Vegetable.Asparagus: {
          break;
        }
      }
    `},
		{Code: `
      enum Vegetable {
        Asparagus = 'asparagus',
      }

      declare const vegetable: Vegetable;

      switch (vegetable) {
        default: {
          break;
        }
      }
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        enum Fruit {
          Apple,
        }
        Fruit.Apple < 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
        }
        Fruit.Apple > 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
        }
        Fruit.Apple == 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
        }
        Fruit.Apple === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
        }
        Fruit.Apple != 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
        }
        Fruit.Apple !== 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple = 0,
          Banana = 'banana',
        }
        Fruit.Apple === 0;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Fruit {
					//     Apple = 0,
					//     Banana = 'banana',
					//   }
					//   Fruit.Apple === Fruit.Apple;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple = 0,
          Banana = 'banana',
        }
        Fruit.Banana === '';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
          Beet = 'beet',
          Celery = 'celery',
        }
        Vegetable.Asparagus === 'beet';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
          Banana,
          Cherry,
        }
        1 === Fruit.Apple;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
          Beet = 'beet',
          Celery = 'celery',
        }
        'beet' === Vegetable.Asparagus;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
          Banana,
          Cherry,
        }
        const fruit = Fruit.Apple;
        fruit === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
          Beet = 'beet',
          Celery = 'celery',
        }
        const vegetable = Vegetable.Asparagus;
        vegetable === 'beet';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
          Banana,
          Cherry,
        }
        const fruit = Fruit.Apple;
        1 === fruit;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
          Beet = 'beet',
          Celery = 'celery',
        }
        const vegetable = Vegetable.Asparagus;
        'beet' === vegetable;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `enum Fruit { Apple, Banana, Cherry }enum Fruit2 {
  Apple2,
  Banana2,
  Cherry2,
}
      Fruit.Apple === Fruit2.Apple2;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
          Beet = 'beet',
          Celery = 'celery',
        }
        enum Vegetable2 {
          Asparagus2 = 'asparagus2',
          Beet2 = 'beet2',
          Celery2 = 'celery2',
        }
        Vegetable.Asparagus === Vegetable2.Asparagus2;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `enum Fruit { Apple, Banana, Cherry }enum Fruit2 {
  Apple2,
  Banana2,
  Cherry2,
}
      const fruit = Fruit.Apple;
      fruit === Fruit2.Apple2;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
          Beet = 'beet',
          Celery = 'celery',
        }
        enum Vegetable2 {
          Asparagus2 = 'asparagus2',
          Beet2 = 'beet2',
          Celery2 = 'celery2',
        }
        const vegetable = Vegetable.Asparagus;
        vegetable === Vegetable2.Asparagus2;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Str {
          A = 'a',
        }
        enum Num {
          B = 1,
        }
        enum Mixed {
          A = 'a',
          B = 1,
        }

        declare const str: Str;
        declare const num: Num;
        declare const mixed: Mixed;

        // following are all errors because the value might be an enum value
        str === 'a';
        num === 1;
        mixed === 'a';
        mixed === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Str {
					//     A = 'a',
					//   }
					//   enum Num {
					//     B = 1,
					//   }
					//   enum Mixed {
					//     A = 'a',
					//     B = 1,
					//   }
					//
					//   declare const str: Str;
					//   declare const num: Num;
					//   declare const mixed: Mixed;
					//
					//   // following are all errors because the value might be an enum value
					//   str === Str.A;
					//   num === 1;
					//   mixed === 'a';
					//   mixed === 1;
					// `,
					//         },
					//       },
				},
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Str {
					//     A = 'a',
					//   }
					//   enum Num {
					//     B = 1,
					//   }
					//   enum Mixed {
					//     A = 'a',
					//     B = 1,
					//   }
					//
					//   declare const str: Str;
					//   declare const num: Num;
					//   declare const mixed: Mixed;
					//
					//   // following are all errors because the value might be an enum value
					//   str === 'a';
					//   num === Num.B;
					//   mixed === 'a';
					//   mixed === 1;
					// `,
					//         },
					//       },
				},
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Str {
					//     A = 'a',
					//   }
					//   enum Num {
					//     B = 1,
					//   }
					//   enum Mixed {
					//     A = 'a',
					//     B = 1,
					//   }
					//
					//   declare const str: Str;
					//   declare const num: Num;
					//   declare const mixed: Mixed;
					//
					//   // following are all errors because the value might be an enum value
					//   str === 'a';
					//   num === 1;
					//   mixed === Mixed.A;
					//   mixed === 1;
					// `,
					//         },
					//       },
				},
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Str {
					//     A = 'a',
					//   }
					//   enum Num {
					//     B = 1,
					//   }
					//   enum Mixed {
					//     A = 'a',
					//     B = 1,
					//   }
					//
					//   declare const str: Str;
					//   declare const num: Num;
					//   declare const mixed: Mixed;
					//
					//   // following are all errors because the value might be an enum value
					//   str === 'a';
					//   num === 1;
					//   mixed === 'a';
					//   mixed === Mixed.B;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple = 'apple',
        }
        type __String =
          | (string & { __escapedIdentifier: void })
          | (void & { __escapedIdentifier: void })
          | Fruit;
        declare const weirdString: __String;
        weirdString === 'someArbitraryValue';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
        }

        declare const fruit: Fruit;

        switch (fruit) {
          case 0: {
            break;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCase",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
          Banana,
        }

        declare const fruit: Fruit;

        switch (fruit) {
          case Fruit.Apple: {
            break;
          }
          case 1: {
            break;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCase",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
        }

        declare const vegetable: Vegetable;

        switch (vegetable) {
          case 'asparagus': {
            break;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCase",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
          Beet = 'beet',
        }

        declare const vegetable: Vegetable;

        switch (vegetable) {
          case Vegetable.Asparagus: {
            break;
          }
          case 'beet': {
            break;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCase",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
          Beet = 'beet',
        }

        declare const vegetable: Vegetable;

        switch (vegetable) {
          case Vegetable.Asparagus: {
            break;
          }
          case 'beet': {
            break;
          }
          default: {
            break;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCase",
				},
			},
		},
		{
			Code: `
        enum Str {
          A = 'a',
          B = 'b',
        }
        declare const str: Str;
        str === 'b';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Str {
					//     A = 'a',
					//     B = 'b',
					//   }
					//   declare const str: Str;
					//   str === Str.B;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Str {
          A = 'a',
          AB = 'ab',
        }
        declare const str: Str;
        str === 'a' + 'b';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Str {
					//     A = 'a',
					//     AB = 'ab',
					//   }
					//   declare const str: Str;
					//   str === Str.AB;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Num {
          A = 1,
          B = 2,
        }
        declare const num: Num;
        1 === num;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Num {
					//     A = 1,
					//     B = 2,
					//   }
					//   declare const num: Num;
					//   Num.A === num;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Num {
          A = 1,
          B = 2,
        }
        declare const num: Num;
        1 /* with */ === /* comment */ num;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Num {
					//     A = 1,
					//     B = 2,
					//   }
					//   declare const num: Num;
					//   Num.A /* with */ === /* comment */ num;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Num {
          A = 1,
          B = 2,
        }
        declare const num: Num;
        1 + 1 === num;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Num {
					//     A = 1,
					//     B = 2,
					//   }
					//   declare const num: Num;
					//   Num.B === num;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Mixed {
          A = 1,
          B = 'b',
        }
        declare const mixed: Mixed;
        mixed === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Mixed {
					//     A = 1,
					//     B = 'b',
					//   }
					//   declare const mixed: Mixed;
					//   mixed === Mixed.A;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Mixed {
          A = 1,
          B = 'b',
        }
        declare const mixed: Mixed;
        mixed === 'b';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum Mixed {
					//     A = 1,
					//     B = 'b',
					//   }
					//   declare const mixed: Mixed;
					//   mixed === Mixed.B;
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum StringKey {
          'test-key' /* with comment */ = 1,
        }
        declare const stringKey: StringKey;
        stringKey === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum StringKey {
					//     'test-key' /* with comment */ = 1,
					//   }
					//   declare const stringKey: StringKey;
					//   stringKey === StringKey['test-key'];
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum StringKey {
          "key-'with-single'-quotes" = 1,
        }
        declare const stringKey: StringKey;
        stringKey === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum StringKey {
					//     "key-'with-single'-quotes" = 1,
					//   }
					//   declare const stringKey: StringKey;
					//   stringKey === StringKey['key-\'with-single\'-quotes'];
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum StringKey {
          'key-"with-double"-quotes' = 1,
        }
        declare const stringKey: StringKey;
        stringKey === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum StringKey {
					//     'key-"with-double"-quotes' = 1,
					//   }
					//   declare const stringKey: StringKey;
					//   stringKey === StringKey['key-"with-double"-quotes'];
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum StringKey {
          'key-` + "`" + `with-backticks` + "`" + `-quotes' = 1,
        }
        declare const stringKey: StringKey;
        stringKey === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum StringKey {
					//     'key-` + "`" + `with-backticks` + "`" + `-quotes' = 1,
					//   }
					//   declare const stringKey: StringKey;
					//   stringKey === StringKey['key-` + "`" + `with-backticks` + "`" + `-quotes'];
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum ComputedKey {
          ['test-key' /* with comment */] = 1,
        }
        declare const computedKey: ComputedKey;
        computedKey === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum ComputedKey {
					//     ['test-key' /* with comment */] = 1,
					//   }
					//   declare const computedKey: ComputedKey;
					//   computedKey === ComputedKey['test-key'];
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum ComputedKey {
          [` + "`" + `test-key` + "`" + ` /* with comment */] = 1,
        }
        declare const computedKey: ComputedKey;
        computedKey === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum ComputedKey {
					//     [` + "`" + `test-key` + "`" + ` /* with comment */] = 1,
					//   }
					//   declare const computedKey: ComputedKey;
					//   computedKey === ComputedKey[` + "`" + `test-key` + "`" + `];
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum ComputedKey {
          [` + "`" + `test-
          key` + "`" + ` /* with comment */] = 1,
        }
        declare const computedKey: ComputedKey;
        computedKey === 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
					//       TODO(port): implement suggestions
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "replaceValueWithEnum",
					//           Output: `
					//   enum ComputedKey {
					//     [` + "`" + `test-
					//     key` + "`" + ` /* with comment */] = 1,
					//   }
					//   declare const computedKey: ComputedKey;
					//   computedKey === ComputedKey[` + "`" + `test-
					//     key` + "`" + `];
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
        }
        declare const foo: number & {};
        if (foo === Fruit.Apple) {
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
        }
        declare const foo: number & { __someBrand: void };
        if (foo === Fruit.Apple) {
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
        }
        declare const foo: string & {};
        if (foo === Vegetable.Asparagus) {
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
		{
			Code: `
        enum Vegetable {
          Asparagus = 'asparagus',
        }
        declare const foo: string & { __someBrand: void };
        if (foo === Vegetable.Asparagus) {
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatchedCondition",
				},
			},
		},
	})
}
