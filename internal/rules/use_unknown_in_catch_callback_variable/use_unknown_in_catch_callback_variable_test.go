package use_unknown_in_catch_callback_variable

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestUseUnknownInCatchCallbackVariable(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &UseUnknownInCatchCallbackVariableRule, []rule_tester.ValidTestCase{
		{Code: `
      Promise.resolve().catch((err: unknown) => {
        throw err;
      });
    `},
		{Code: `
      let x = Math.random() ? 'ca' + 'tch' : 'catch';
      Promise.resolve()[x]((err: Error) => {});
    `},
		{Code: `
      Promise.resolve().then(
        () => {},
        (err: unknown) => {
          throw err;
        },
      );
    `},
		{Code: `
      Promise.resolve().catch(() => {
        throw new Error();
      });
    `},
		{Code: `
      Promise.reject(new Error()).catch("not this rule's problem");
    `},
		{Code: `
      declare const crappyHandler: (() => void) | 2;
      Promise.reject(new Error()).catch(crappyHandler);
    `},
		{Code: `
      Promise.resolve().catch((...args: [unknown]) => {
        throw args[0];
      });
    `},
		{Code: `
      Promise.resolve().catch((...args: [a: unknown]) => {
        const err = args[0];
      });
    `},
		{Code: `
      Promise.resolve().catch((...args: readonly unknown[]) => {
        throw args[0];
      });
    `},
		{Code: `
      declare const notAPromise: { catch: (f: Function) => void };
      notAPromise.catch((...args: [a: string, string]) => {
        throw args[0];
      });
    `},
		{Code: `
      declare const catchArgs: [(x: unknown) => void];
      Promise.reject(new Error()).catch(...catchArgs);
    `},
		{Code: `
      declare const catchArgs: [
        string | (() => never),
        (shouldntFlag: string) => void,
        number,
      ];
      Promise.reject(new Error()).catch(...catchArgs);
    `},
		{Code: `
      declare const catchArgs: ['not callable'];
      Promise.reject(new Error()).catch(...catchArgs);
    `},
		{Code: `
      declare const emptySpread: [];
      Promise.reject(new Error()).catch(...emptySpread);
    `},
		{Code: `
      Promise.resolve().catch(
        (
          ...err: [unknown, string | ((number | unknown) & { b: () => void }), string]
        ) => {
          throw err;
        },
      );
    `},
		{Code: `
      declare const notAMemberExpression: (...args: any[]) => {};
      notAMemberExpression(
        'This helps get 100% code cov',
        "but doesn't test anything useful related to the rule.",
      );
    `},
		{Code: `
      Promise.resolve().catch((...[args]: [unknown]) => {
        console.log(args);
      });
    `},
		{Code: `
      Promise.resolve().catch((...{ find }: [unknown]) => {
        console.log(find);
      });
    `},
		{Code: "Promise.resolve.then();"},
		{Code: "Promise.resolve().then(() => {});"},
		{Code: `
      declare const singleTupleArg: [() => void];
      Promise.resolve().then(...singleTupleArg, (error: unknown) => {});
    `},
		{Code: `
      declare const arrayArg: (() => void)[];
      Promise.resolve().then(...arrayArg, error => {});
    `},
		{Code: `
declare let iPromiseImAPromise: Promise<any>;
declare const catchArgs: [(x: any) => void];
iPromiseImAPromise.catch(...catchArgs);
    `},
		{Code: `
declare const catchArgs: [
  string | (() => never) | ((x: string) => void),
  number,
];
Promise.reject(new Error()).catch(...catchArgs);
    `},
		{Code: `
declare const you: [];
declare const cannot: [];
declare const fool: [];
declare const me: [(x: Error) => void] | undefined;
Promise.resolve(undefined).catch(...you, ...cannot, ...fool, ...me!);
    `},
		{Code: `
declare const really: undefined[];
declare const dumb: [];
declare const code: (x: Error) => void;
Promise.resolve(undefined).catch(...really, ...dumb, code);
    `},
		{Code: `
declare const x: ((x: any) => string)[];
Promise.resolve('string promise').catch(...x);
    `},
		{Code: `
declare const x: any;
Promise.resolve().catch(...x);
    `},
		{Code: `
declare const thenArgs: [() => {}, (err: any) => {}];
Promise.resolve().then(...thenArgs);
    `},
		{Code: `
declare const yoloHandler: (x: any) => void;
Promise.reject(new Error('I will reject!')).catch(yoloHandler);
    `},
		{Code: `
type InvalidHandler = (arg: any) => void;
Promise.resolve().catch(<InvalidHandler>(
  function (err /* awkward spot for comment */) {
    throw err;
  }
));
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
Promise.resolve().catch((err: Error) => {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((err: unknown) => {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        let method = 'catch';
        Promise.resolve()[method]((error: Error) => {});
      `,
			// TODO(port): method's type is string. So getAccessedPropertyName doesn't support it. Do we even need to support such cases?
			Skip: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
        let method = 'catch';
        Promise.resolve()[method]((error: unknown) => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch((e, ...rest: []) => {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((e: unknown, ...rest: []) => {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch(
  (err: string | ((number | unknown) & { b: () => void })) => {
    throw err;
  },
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch(
  (err: unknown) => {
    throw err;
  },
);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch(
  (
    ...err: [
      unknown[],
      string | ((number | unknown) & { b: () => void }),
      string,
    ]
  ) => {
    throw err;
  },
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongRestTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch(
  (
    ...err: [unknown]
  ) => {
    throw err;
  },
);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch(function (err: string) {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch(function (err: unknown) {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch(function (err /* awkward spot for comment */) {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch(function (err: unknown /* awkward spot for comment */) {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch(function namedCallback(err: string) {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch(function namedCallback(err: unknown) {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch(err => {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((err: unknown) => {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().then(
  () => {},
  error => {},
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
Promise.resolve().then(
  () => {},
  (error: unknown) => {},
);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch((err?) => {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((err?: unknown) => {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch((err?: string) => {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((err?: unknown) => {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch(err/* with comment */=> {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((err: unknown)/* with comment */=> {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch((err = 2) => {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((err: unknown = 2) => {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch((err: any /* comment 1 */ = /* comment 2 */ 2) => {
  throw err;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((err: unknown /* comment 1 */ = /* comment 2 */ 2) => {
  throw err;
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch((...args) => {
  throw args[0];
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownRestTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((...args: [unknown]) => {
  throw args[0];
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject(new Error('I will reject!')).catch(([err]: [unknown]) => {
  console.log(err);
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknownArrayDestructuringPattern",
					Line:      2,
				},
			},
		},
		{
			Code: `
Promise.resolve(' a string ').catch(
  (a: any, b: () => any, c: (x: string & number) => void) => {},
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
Promise.resolve(' a string ').catch(
  (a: unknown, b: () => any, c: (x: string & number) => void) => {},
);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve('object destructuring').catch(({}) => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknownObjectDestructuringPattern",
					Line:      2,
				},
			},
		},
		{
			Code: `
Promise.resolve('object destructuring').catch(function ({ gotcha }) {
  return null;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknownObjectDestructuringPattern",
					Line:      2,
				},
			},
		},
		{
			Code: `
Promise.resolve()['catch']((x: any) => 'return');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongTypeAnnotationSuggestion",
							Output: `
Promise.resolve()['catch']((x: unknown) => 'return');
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject().catch((...x: any) => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongRestTypeAnnotationSuggestion",
							Output: `
Promise.reject().catch((...x: [unknown]) => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch((...[args]: [string]) => {
  console.log(args);
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongRestTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((...[args]: [unknown]) => {
  console.log(args);
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().catch((...{ find }: [string]) => {
  console.log(find);
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrongRestTypeAnnotationSuggestion",
							Output: `
Promise.resolve().catch((...{ find }: [unknown]) => {
  console.log(find);
});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const condition: boolean;
Promise.resolve('foo').then(() => {}, condition ? err => {} : err => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      3,
					Column:    51,
					EndLine:   3,
					EndColumn: 54,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
declare const condition: boolean;
Promise.resolve('foo').then(() => {}, condition ? (err: unknown) => {} : err => {});
      `,
						},
					},
				},
				{
					MessageId: "useUnknown",
					Line:      3,
					Column:    63,
					EndLine:   3,
					EndColumn: 66,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
declare const condition: boolean;
Promise.resolve('foo').then(() => {}, condition ? err => {} : (err: unknown) => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const condition: boolean;
declare const maybeNullishHandler: null | ((err: any) => void);
Promise.resolve('foo').catch(
  condition
    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))
    : (condition && (err => {})) || (err => {}),
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useUnknown",
					Line:      6,
					Column:    56,
					EndLine:   6,
					EndColumn: 59,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
declare const condition: boolean;
declare const maybeNullishHandler: null | ((err: any) => void);
Promise.resolve('foo').catch(
  condition
    ? ((err => {}, err => {}, maybeNullishHandler) ?? ((err: unknown) => {}))
    : (condition && (err => {})) || (err => {}),
);
      `,
						},
					},
				},
				{
					MessageId: "useUnknown",
					Line:      7,
					Column:    22,
					EndLine:   7,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
declare const condition: boolean;
declare const maybeNullishHandler: null | ((err: any) => void);
Promise.resolve('foo').catch(
  condition
    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))
    : (condition && ((err: unknown) => {})) || (err => {}),
);
      `,
						},
					},
				},
				{
					MessageId: "useUnknown",
					Line:      7,
					Column:    38,
					EndLine:   7,
					EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addUnknownTypeAnnotationSuggestion",
							Output: `
declare const condition: boolean;
declare const maybeNullishHandler: null | ((err: any) => void);
Promise.resolve('foo').catch(
  condition
    ? ((err => {}, err => {}, maybeNullishHandler) ?? (err => {}))
    : (condition && (err => {})) || ((err: unknown) => {}),
);
      `,
						},
					},
				},
			},
		},
	})
}
