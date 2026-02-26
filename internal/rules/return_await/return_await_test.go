package return_await

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestReturnAwait(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &ReturnAwaitRule, []rule_tester.ValidTestCase{
		{Code: "return;"},
		{Code: `
      function test() {
        return;
      }
    `},
		{Code: `
      function test() {
        return 1;
      }
    `},
		{Code: `
      async function test() {
        return;
      }
    `},
		{Code: `
      async function test() {
        return 1;
      }
    `},
		{Code: "const test = () => 1;"},
		{Code: "const test = async () => 1;"},
		{Code: `
      async function test() {
        return Promise.resolve(1);
      }
    `},
		{Code: `
      async function test() {
        try {
          return await Promise.resolve(1);
        } catch (e) {
          return await Promise.resolve(2);
        } finally {
          console.log('cleanup');
        }
      }
    `},
		{Code: `
const fn = (): any => null;
async function test() {
  return await fn();
}
    `},
		{Code: `
const fn = (): unknown => null;
async function test() {
  return await fn();
}
    `},
		{Code: `
async function test(unknownParam: unknown) {
  try {
    return await unknownParam;
  } finally {
    console.log('In finally block');
  }
}
    `},
		{
			Code: `
        async function test() {
          if (Math.random() < 0.33) {
            return await Promise.resolve(1);
          } else if (Math.random() < 0.5) {
            return Promise.resolve(2);
          }

          try {
          } catch (e) {
            return await Promise.resolve(3);
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"error-handling-correctness-only"`),
		},
		{Code: `
      async function test() {
        try {
          const one = await Promise.resolve(1);
          return one;
        } catch (e) {
          const two = await Promise.resolve(2);
          return two;
        } finally {
          console.log('cleanup');
        }
      }
    `},
		{
			Code: `
        function test() {
          return 1;
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
        async function test() {
          return 1;
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code:    "const test = () => 1;",
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code:    "const test = async () => 1;",
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
        async function test() {
          return Promise.resolve(1);
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
        async function test() {
          try {
            return await Promise.resolve(1);
          } catch (e) {
            return await Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
        async function test() {
          try {
            throw 'foo';
          } catch (e) {
            return Promise.resolve(1);
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
        async function test() {
          try {
            throw 'foo';
          } catch (e) {
            throw 'foo2';
          } finally {
            return Promise.resolve(1);
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
        async function test() {
          try {
            const one = await Promise.resolve(1);
            return one;
          } catch (e) {
            const two = await Promise.resolve(2);
            return two;
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
        async function test() {
          return Promise.resolve(1);
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"never"`),
		},
		{
			Code:    "const test = async () => Promise.resolve(1);",
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"never"`),
		},
		{
			Code: `
        async function test() {
          try {
            return Promise.resolve(1);
          } catch (e) {
            return Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"never"`),
		},
		{
			Code: `
        async function test() {
          return await Promise.resolve(1);
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
		},
		{
			Code:    "const test = async () => await Promise.resolve(1);",
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
		},
		{
			Code:    "const f = (): Promise<number> => Promise.resolve(1);",
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
		},
		{
			Code:    "const f = (): Promise<number> => Promise.resolve(1);",
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code:    "const f = (): Promise<number> => Promise.resolve(1);",
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"error-handling-correctness-only"`),
		},
		{
			Code: `
        async function test() {
          try {
            return await Promise.resolve(1);
          } catch (e) {
            return await Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
		},
		{
			Code: `
        declare function foo(): Promise<boolean>;

        function bar(baz: boolean): Promise<boolean> | boolean {
          if (baz) {
            return true;
          } else {
            return foo();
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
		},
		{
			Code: `
        async function test(): Promise<string> {
          const res = await Promise.resolve('{}');
          try {
            return JSON.parse(res);
          } catch (error) {
            return res;
          }
        }
      `,
		},
		{
			Code: `
        async function test() {
          const res = await Promise.resolve('{}');
          try {
            async function nested() {
              return Promise.resolve('ok');
            }
            return await nested();
          } catch (error) {
            return Promise.resolve('error');
          }
        }
      `,
		},
		{
			Code: `
async function f() {
  try {
  } catch {
    try {
    } catch {
      return Promise.reject();
    }
  }
}
      `,
		},
		{
			Code: `
async function f() {
  try {
  } finally {
    try {
    } catch {
      return Promise.reject();
    }
  }
}
      `,
		},
		{
			Code: `
async function f() {
  try {
  } finally {
    try {
    } finally {
      try {
      } catch {
        return Promise.reject();
      }
    }
  }
}
      `,
		},
		{
			Code: `
declare const bleh: any;
async function f() {
  using something = bleh;
  return await Promise.resolve(2);
}
      `,
		},
		{
			Code: `
declare const bleh: any;
async function f() {
  await using something = bleh;
  return await Promise.resolve(2);
}
      `,
		},
		{
			Code: `
declare const bleh: any;
async function f() {
  using something = bleh;
  {
    return await Promise.resolve(2);
  }
}
      `,
		},
		{
			Code: `
declare const bleh: any;
async function f() {
  return Promise.resolve(2);
  using something = bleh;
}
      `,
		},
		{
			Code: `
declare const bleh: any;
async function f() {
  return await Promise.resolve(2);
  using something = bleh;
}
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
		},
		{
			Code: `
declare function asyncFn(): Promise<unknown>;
async function returnAwait() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  return await asyncFn();
}
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  async function innerFunction() {
    return asyncFn();
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  const innerFunction = async () => asyncFn();
}
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
		},
		{
			Code: `
using foo = 1 as any;
return Promise.resolve(42);
      `,
		},
		{
			Code: `
{
  using foo = 1 as any;
  return Promise.resolve(42);
}
      `,
		},
		{
			Code: `
async function wrapper<T>(value: T) {
  return await value;
}
      `,
		},
		{
			Code: `
async function wrapper<T extends unknown>(value: T) {
  return await value;
}
      `,
		},
		{
			Code: `
async function wrapper<T extends any>(value: T) {
  return await value;
}
      `,
		},
		{
			Code: `
class C<T> {
  async wrapper<T>(value: T) {
    return await value;
  }
}
      `,
		},
		{
			Code: `
class C<R> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `,
		},
		{
			Code: `
class C<R extends unknown> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        async function test() {
          return await 1;
        }
      `,
			Output: []string{`
        async function test() {
          return  1;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code: `
        async function test() {
          const foo = 1;
          return await { foo };
        }
      `,
			Output: []string{`
        async function test() {
          const foo = 1;
          return  { foo };
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      4,
				},
			},
		},
		{
			Code: `
        async function test() {
          const foo = 1;
          return await foo;
        }
      `,
			Output: []string{`
        async function test() {
          const foo = 1;
          return  foo;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      4,
				},
			},
		},
		{
			Code:   "const test = async () => await 1;",
			Output: []string{"const test = async () =>  1;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      1,
				},
			},
		},
		{
			Code:   "const test = async () => await /* comment */ 1;",
			Output: []string{"const test = async () =>  /* comment */ 1;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      1,
				},
			},
		},
		{
			Code:   "const test = async () => await Promise.resolve(1);",
			Output: []string{"const test = async () =>  Promise.resolve(1);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      1,
				},
			},
		},
		{
			Code: `
        async function test() {
          try {
            return Promise.resolve(1);
          } catch (e) {
            return Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"error-handling-correctness-only"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            return await Promise.resolve(1);
          } catch (e) {
            return Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
						},
					},
				},
				{
					MessageId: "requiredPromiseAwait",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            return Promise.resolve(1);
          } catch (e) {
            return await Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        async function test() {
          try {
            return Promise.resolve(1);
          } catch (e) {
            return Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            return await Promise.resolve(1);
          } catch (e) {
            return Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
						},
					},
				},
				{
					MessageId: "requiredPromiseAwait",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            return Promise.resolve(1);
          } catch (e) {
            return await Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        async function test() {
          try {
            return Promise.resolve(1);
          } catch (e) {
            return Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            return await Promise.resolve(1);
          } catch (e) {
            return Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
						},
					},
				},
				{
					MessageId: "requiredPromiseAwait",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            return Promise.resolve(1);
          } catch (e) {
            return await Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        async function test() {
          return await Promise.resolve(1);
        }
      `,
			Output: []string{`
        async function test() {
          return  Promise.resolve(1);
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code: `
        async function test() {
          return await 1;
        }
      `,
			Output: []string{`
        async function test() {
          return  1;
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code:    "const test = async () => await 1;",
			Output:  []string{"const test = async () =>  1;"},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      1,
				},
			},
		},
		{
			Code:    "const test = async () => await Promise.resolve(1);",
			Output:  []string{"const test = async () =>  Promise.resolve(1);"},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      1,
				},
			},
		},
		{
			Code: `
        async function test() {
          return await Promise.resolve(1);
        }
      `,
			Output: []string{`
        async function test() {
          return  Promise.resolve(1);
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code: `
        async function test() {
          return await 1;
        }
      `,
			Output: []string{`
        async function test() {
          return  1;
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"never"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code: `
        async function test() {
          try {
            return await Promise.resolve(1);
          } catch (e) {
            return await Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"never"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "disallowedPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            return  Promise.resolve(1);
          } catch (e) {
            return await Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
						},
					},
				},
				{
					MessageId: "disallowedPromiseAwait",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "disallowedPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            return await Promise.resolve(1);
          } catch (e) {
            return  Promise.resolve(2);
          } finally {
            console.log('cleanup');
          }
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        async function test() {
          return await Promise.resolve(1);
        }
      `,
			Output: []string{`
        async function test() {
          return  Promise.resolve(1);
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"never"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code: `
        async function test() {
          return await 1;
        }
      `,
			Output: []string{`
        async function test() {
          return  1;
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code: `
        async function test() {
          return Promise.resolve(1);
        }
      `,
			Output: []string{`
        async function test() {
          return await Promise.resolve(1);
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code:    "const test = async () => Promise.resolve(1);",
			Output:  []string{"const test = async () => await Promise.resolve(1);"},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      1,
				},
			},
		},
		{
			Code: `
async function foo() {}
async function bar() {}
async function baz() {}
async function qux() {}
async function buzz() {
  return (await foo()) ? bar() : baz();
}
      `,
			Output: []string{`
async function foo() {}
async function bar() {}
async function baz() {}
async function qux() {}
async function buzz() {
  return (await foo()) ? await bar() : await baz();
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      7,
				},
				{
					MessageId: "requiredPromiseAwait",
					Line:      7,
				},
			},
		},
		{
			Code: `
async function foo() {}
async function bar() {}
async function baz() {}
async function qux() {}
async function buzz() {
  return (await foo())
    ? (
      bar ? bar() : baz()
    ) : baz ? baz() : bar();
}
      `,
			Output: []string{`
async function foo() {}
async function bar() {}
async function baz() {}
async function qux() {}
async function buzz() {
  return (await foo())
    ? (
      bar ? await bar() : await baz()
    ) : baz ? await baz() : await bar();
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      10,
				},
				{
					MessageId: "requiredPromiseAwait",
					Line:      10,
				},
				{
					MessageId: "requiredPromiseAwait",
					Line:      9,
				},
				{
					MessageId: "requiredPromiseAwait",
					Line:      9,
				},
			},
		},
		{
			Code: `
async function foo() {}
async function bar() {}
async function buzz() {
  return (await foo()) ? await 1 : bar();
}
      `,
			Output: []string{`
async function foo() {}
async function bar() {}
async function buzz() {
  return (await foo()) ?  1 : await bar();
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      5,
				},
				{
					MessageId: "nonPromiseAwait",
					Line:      5,
				},
			},
		},
		{
			Code: `
async function foo() {}
async function bar() {}
async function baz() {}
const buzz = async () => ((await foo()) ? bar() : baz());
      `,
			Output: []string{`
async function foo() {}
async function bar() {}
async function baz() {}
const buzz = async () => ((await foo()) ? await bar() : await baz());
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      5,
				},
				{
					MessageId: "requiredPromiseAwait",
					Line:      5,
				},
			},
		},
		{
			Code: `
async function foo() {}
async function bar() {}
const buzz = async () => ((await foo()) ? await 1 : bar());
      `,
			Output: []string{`
async function foo() {}
async function bar() {}
const buzz = async () => ((await foo()) ?  1 : await bar());
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      4,
				},
				{
					MessageId: "nonPromiseAwait",
					Line:      4,
				},
			},
		},
		{
			Code: `
async function test<T>(): Promise<T> {
  const res = await fetch('...');
  try {
    return res.json() as Promise<T>;
  } catch (err) {
    throw Error('Request Failed.');
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
async function test<T>(): Promise<T> {
  const res = await fetch('...');
  try {
    return await (res.json() as Promise<T>);
  } catch (err) {
    throw Error('Request Failed.');
  }
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        async function test() {
          try {
            const callback1 = function () {};
            const callback2 = async function () {};
            function callback3() {}
            async function callback4() {}
            const callback5 = () => {};
            const callback6 = async () => {};
            return Promise.resolve('try');
          } finally {
            return Promise.resolve('finally');
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function test() {
          try {
            const callback1 = function () {};
            const callback2 = async function () {};
            function callback3() {}
            async function callback4() {}
            const callback5 = () => {};
            const callback6 = async () => {};
            return await Promise.resolve('try');
          } finally {
            return Promise.resolve('finally');
          }
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        async function bar() {}
        async function foo() {
          try {
            return undefined || bar();
          } catch {}
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function bar() {}
        async function foo() {
          try {
            return await (undefined || bar());
          } catch {}
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        async function bar() {}
        async function foo() {
          try {
            return bar() || undefined || bar();
          } catch {}
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        async function bar() {}
        async function foo() {
          try {
            return await (bar() || undefined || bar());
          } catch {}
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function bar() {}
async function func1() {
  try {
    return null ?? bar();
  } catch {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
async function bar() {}
async function func1() {
  try {
    return await (null ?? bar());
  } catch {}
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function bar() {}
async function func2() {
  try {
    return 1 && bar();
  } catch {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
async function bar() {}
async function func2() {
  try {
    return await (1 && bar());
  } catch {}
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
const foo = {
  bar: async function () {},
};
async function func3() {
  try {
    return foo.bar();
  } catch {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
const foo = {
  bar: async function () {},
};
async function func3() {
  try {
    return await foo.bar();
  } catch {}
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        class X {
          async bar() {
            return;
          }
          async func2() {
            try {
              return this.bar();
            } catch {}
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
        class X {
          async bar() {
            return;
          }
          async func2() {
            try {
              return await this.bar();
            } catch {}
          }
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        async function test() {
          const res = await Promise.resolve('{}');
          try {
            async function nested() {
              return Promise.resolve('ok');
            }
            return await nested();
          } catch (error) {
            return await Promise.resolve('error');
          }
        }
      `,
			Output: []string{`
        async function test() {
          const res = await Promise.resolve('{}');
          try {
            async function nested() {
              return Promise.resolve('ok');
            }
            return await nested();
          } catch (error) {
            return  Promise.resolve('error');
          }
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      10,
				},
			},
		},
		{
			Code: `
async function f() {
  try {
    try {
    } finally {
      // affects error handling of outer catch
      return Promise.reject();
    }
  } catch {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
async function f() {
  try {
    try {
    } finally {
      // affects error handling of outer catch
      return await Promise.reject();
    }
  } catch {}
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function f() {
  try {
    try {
    } catch {
      // affects error handling of outer catch
      return Promise.reject();
    }
  } catch {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
async function f() {
  try {
    try {
    } catch {
      // affects error handling of outer catch
      return await Promise.reject();
    }
  } catch {}
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function f() {
  try {
  } catch {
    try {
    } finally {
      try {
      } catch {
        return Promise.reject();
      }
    }
  } finally {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
async function f() {
  try {
  } catch {
    try {
    } finally {
      try {
      } catch {
        return await Promise.reject();
      }
    }
  } finally {
  }
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const bleh: any;
async function f() {
  if (cond) {
    using something = bleh;
    if (anotherCondition) {
      return Promise.resolve(2);
    }
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
declare const bleh: any;
async function f() {
  if (cond) {
    using something = bleh;
    if (anotherCondition) {
      return await Promise.resolve(2);
    }
  }
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const bleh: any;
async function f() {
  if (cond) {
    await using something = bleh;
    if (anotherCondition) {
      return Promise.resolve(2);
    }
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "requiredPromiseAwaitSuggestion",
							Output: `
declare const bleh: any;
async function f() {
  if (cond) {
    await using something = bleh;
    if (anotherCondition) {
      return await Promise.resolve(2);
    }
  }
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const bleh: any;
async function f() {
  if (cond) {
    using something = bleh;
  } else if (anotherCondition) {
    return Promise.resolve(2);
  }
}
      `,
			Output: []string{`
declare const bleh: any;
async function f() {
  if (cond) {
    using something = bleh;
  } else if (anotherCondition) {
    return await Promise.resolve(2);
  }
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"always"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requiredPromiseAwait",
					Line:      7,
				},
			},
		},
		{
			Code: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  async function innerFunction() {
    return await asyncFn();
  }
}
      `,
			Output: []string{`
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  async function innerFunction() {
    return  asyncFn();
  }
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      11,
				},
			},
		},
		{
			Code: `
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  const innerFunction = async () => await asyncFn();
}
      `,
			Output: []string{`
declare function asyncFn(): Promise<unknown>;
async function outerFunction() {
  using _ = {
    [Symbol.dispose]: () => {
      console.log('dispose');
    },
  };

  const innerFunction = async () =>  asyncFn();
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[ReturnAwaitOptions](`"in-try-catch"`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "disallowedPromiseAwait",
					Line:      10,
				},
			},
		},
		{
			Code: `
async function wrapper<T extends number>(value: T) {
  return await value;
}
      `,
			Output: []string{`
async function wrapper<T extends number>(value: T) {
  return  value;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      3,
				},
			},
		},
		{
			Code: `
class C<T> {
  async wrapper<T extends string>(value: T) {
    return await value;
  }
}
      `,
			Output: []string{`
class C<T> {
  async wrapper<T extends string>(value: T) {
    return  value;
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      4,
				},
			},
		},
		{
			Code: `
class C<R extends number> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `,
			Output: []string{`
class C<R extends number> {
  async wrapper<T extends R>(value: T) {
    return  value;
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "nonPromiseAwait",
					Line:      4,
				},
			},
		},
	})
}
