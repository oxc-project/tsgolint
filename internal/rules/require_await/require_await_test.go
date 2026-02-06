package require_await

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestRequireAwaitRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &RequireAwaitRule, []rule_tester.ValidTestCase{
		{Code: `
function numberOne(): number {
  return 1;
}
    `},
		{Code: `
const numberOne = function (): number {
  return 1;
};
    `},
		{Code: `
      const numberOne = (): number => 1;
    `},
		{Code: `
const numberOne = (): number => {
  return 1;
};
    `},
		{Code: `
function delay() {
  return Promise.resolve();
}
    `},
		{Code: `
const delay = () => {
  return Promise.resolve();
};
    `},
		{Code: "const delay = () => Promise.resolve();"},
		{Code: `
async function numberOne(): Promise<number> {
  return await 1;
}
    `},
		{Code: `
const numberOne = async function (): Promise<number> {
  return await 1;
};
    `},
		{Code: "const numberOne = async (): Promise<number> => await 1;"},
		{Code: `
const numberOne = async (): Promise<number> => {
  return await 1;
};
    `},
		{Code: `
async function numberOne(): Promise<number> {
  return Promise.resolve(1);
}
    `},
		{Code: `
const numberOne = async function (): Promise<number> {
  return Promise.resolve(1);
};
    `},
		{Code: "const numberOne = async (): Promise<number> => Promise.resolve(1);"},
		{Code: `
const numberOne = async (): Promise<number> => {
  return Promise.resolve(1);
};
    `},
		{Code: `
async function numberOne(): Promise<number> {
  return getAsyncNumber(1);
}
async function getAsyncNumber(x: number): Promise<number> {
  return Promise.resolve(x);
}
    `},
		{Code: `
const numberOne = async function (): Promise<number> {
  return getAsyncNumber(1);
};
const getAsyncNumber = async function (x: number): Promise<number> {
  return Promise.resolve(x);
};
    `},
		{Code: `
const numberOne = async (): Promise<number> => getAsyncNumber(1);
const getAsyncNumber = async function (x: number): Promise<number> {
  return Promise.resolve(x);
};
    `},
		{Code: `
const numberOne = async (): Promise<number> => {
  return getAsyncNumber(1);
};
const getAsyncNumber = async function (x: number): Promise<number> {
  return Promise.resolve(x);
};
    `},
		{Code: `
async function testFunction(): Promise<void> {
  await Promise.all(
    [1, 2, 3].map(
      // this should not trigger an error on the parent function
      async value => Promise.resolve(value),
    ),
  );
}
    `},
		{Code: `
function* test6() {
  yield* syncGenerator();
}
    `},
		{Code: `
function* syncGenerator() {
  yield 1;
}
    `},
		{Code: `
async function* asyncGenerator() {
  await Promise.resolve();
  yield 1;
}
async function* test1() {
  yield* asyncGenerator();
}
    `},
		{Code: `
async function* asyncGenerator() {
  await Promise.resolve();
  yield 1;
}
async function* test1() {
  yield* asyncGenerator();
  yield* 2;
}
    `},
		{Code: `
async function* test(source: AsyncIterable<any>) {
  yield* source;
}
    `},
		{Code: `
async function* test(source: Iterable<any> & AsyncIterable<any>) {
  yield* source;
}
    `},
		{Code: `
async function* test(source: Iterable<any> | AsyncIterable<any>) {
  yield* source;
}
    `},
		{Code: `
type MyType = {
  [Symbol.iterator](): Iterator<any>;
  [Symbol.asyncIterator](): AsyncIterator<any>;
};
async function* test(source: MyType) {
  yield* source;
}
    `},
		{Code: `
type MyType = {
  [Symbol.asyncIterator]: () => AsyncIterator<any>;
};
async function* test(source: MyType) {
  yield* source;
}
    `},
		{Code: `
type MyFunctionType = () => AsyncIterator<any>;
type MyType = {
  [Symbol.asyncIterator]: MyFunctionType;
};
async function* test(source: MyType) {
  yield* source;
}
    `},
		{Code: `
async function* foo(): Promise<string> {
  return new Promise(res => res(` + "`" + `hello` + "`" + `));
}
    `},
		{Code: `
      async function* f() {
        let x!: Omit<
          {
            [Symbol.asyncIterator](): AsyncIterator<any>;
          },
          'z'
        >;
        yield* x;
      }
    `},
		{Code: `
      const fn = async () => {
        await using foo = new Bar();
      };
    `},
		{Code: `
      async function* test1() {
        yield Promise.resolve(1);
      }
    `},
		{Code: `
      function asyncFunction() {
        return Promise.resolve(1);
      }
      async function* test1() {
        yield asyncFunction();
      }
    `},
		{Code: `
      declare const asyncFunction: () => Promise<void>;
      async function* test1() {
        yield asyncFunction();
      }
    `},
		{Code: `
      async function* test1() {
        yield new Promise(() => {});
      }
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
async function numberOne(): Promise<number> {
  return 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function numberOne(): number {
					//   return 1;
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
const numberOne = async function (): Promise<number> {
  return 1;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// const numberOne = function (): number {
					//   return 1;
					// };
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: "const numberOne = async (): Promise<number> => 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//   {
					//     MessageId: "removeAsync",
					//     Output: "const numberOne = (): number => 1;",
					//   },
					// },
				},
			},
		},
		{
			Code: `
async function values(): Promise<Array<number>> {
  return [1];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function values(): Array<number> {
					//   return [1];
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
        async function foo() {
          function nested() {
            await doSomething();
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function foo() {
					//     function nested() {
					//       await doSomething();
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
async function* foo(): void {
  doSomething();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* foo(): void {
					//   doSomething();
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
async function* foo() {
  yield 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* foo() {
					//   yield 1;
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
const foo = async function* () {
  console.log('bar');
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// const foo = function* () {
					//   console.log('bar');
					// };
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
async function* asyncGenerator() {
  yield 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* asyncGenerator() {
					//   yield 1;
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
async function* asyncGenerator(source: Iterable<any>) {
  yield* source;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* asyncGenerator(source: Iterable<any>) {
					//   yield* source;
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
function isAsyncIterable(value: unknown): value is AsyncIterable<any> {
  return true;
}
async function* asyncGenerator(source: Iterable<any> | AsyncIterable<any>) {
  if (!isAsyncIterable(source)) {
    yield* source;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function isAsyncIterable(value: unknown): value is AsyncIterable<any> {
					//   return true;
					// }
					// function* asyncGenerator(source: Iterable<any> | AsyncIterable<any>) {
					//   if (!isAsyncIterable(source)) {
					//     yield* source;
					//   }
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
function* syncGenerator() {
  yield 1;
}
async function* asyncGenerator() {
  yield* syncGenerator();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* syncGenerator() {
					//   yield 1;
					// }
					// function* asyncGenerator() {
					//   yield* syncGenerator();
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
async function* asyncGenerator() {
  yield* anotherAsyncGenerator(); // Unknown function.
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* asyncGenerator() {
					//   yield* anotherAsyncGenerator(); // Unknown function.
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
        const fn = async () => {
          using foo = new Bar();
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   const fn = () => {
					//     using foo = new Bar();
					//   };
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        // intentional TS error
        async function* foo(): Promise<number> {
          yield 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   // intentional TS error
					//   function* foo(): Promise<number> {
					//     yield 1;
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async function* foo(): AsyncGenerator {
          yield 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function* foo(): Generator {
					//     yield 1;
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async function* foo(): AsyncGenerator<number> {
          yield 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function* foo(): Generator<number> {
					//     yield 1;
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
	})
}

func TestRequireAwaitRuleEslintBase(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &RequireAwaitRule, []rule_tester.ValidTestCase{
		{Code: `
async function foo() {
  await doSomething();
}
    `},
		{Code: `
(async function () {
  await doSomething();
});
    `},
		{Code: `
async () => {
  await doSomething();
};
    `},
		{Code: "async () => await doSomething();"},
		{Code: `
({
  async foo() {
    await doSomething();
  },
});
    `},
		{Code: `
class A {
  async foo() {
    await doSomething();
  }
}
    `},
		{Code: `
(class {
  async foo() {
    await doSomething();
  }
});
    `},
		{Code: `
async function foo() {
  await (async () => {
    await doSomething();
  });
}
    `},
		{Code: "async function foo() {}"},
		{Code: "async () => {};"},
		{Code: `
function foo() {
  doSomething();
}
    `},
		{Code: `
async function foo() {
  for await (x of xs);
}
    `},
		{
			Code: "await foo();",
		},
		{
			Code: `
for await (let num of asyncIterable) {
  console.log(num);
}
      `,
		},
		{
			Code: `
        async function* run() {
          await new Promise(resolve => setTimeout(resolve, 100));
          yield 'Hello';
          console.log('World');
        }
      `,
		},
		{
			Code: "async function* run() {}",
		},
		{
			Code: "const foo = async function* () {};",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        async function foo() {
          doSomething();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function foo() {
					//     doSomething();
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        (async function () {
          doSomething();
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   (function () {
					//     doSomething();
					//   });
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async () => {
          doSomething();
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   () => {
					//     doSomething();
					//   };
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: "async () => doSomething();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//   {
					//     MessageId: "removeAsync",
					//     Output: "() => doSomething();",
					//   },
					// },
				},
			},
		},
		{
			Code: `
        ({
          async foo() {
            doSomething();
          },
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   ({
					//     foo() {
					//       doSomething();
					//     },
					//   });
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        class A {
          async foo() {
            doSomething();
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   class A {
					//     foo() {
					//       doSomething();
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        class A {
          public async foo() {
            doSomething();
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   class A {
					//     public foo() {
					//       doSomething();
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        (class {
          async foo() {
            doSomething();
          }
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   (class {
					//     foo() {
					//       doSomething();
					//     }
					//   });
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        (class {
          async ''() {
            doSomething();
          }
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   (class {
					//     ''() {
					//       doSomething();
					//     }
					//   });
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async function foo() {
          async () => {
            await doSomething();
          };
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function foo() {
					//     async () => {
					//       await doSomething();
					//     };
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async function foo() {
          await (async () => {
            doSomething();
          });
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   async function foo() {
					//     await (() => {
					//       doSomething();
					//     });
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        const obj = {
          async: async function foo() {
            bar();
          },
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   const obj = {
					//     async: function foo() {
					//       bar();
					//     },
					//   };
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async    /* test */ function foo() {
          doSomething();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   /* test */ function foo() {
					//     doSomething();
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        class A {
          a = 0
          async [b]() {
            return 0;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   class A {
					//     a = 0
					//     ;[b]() {
					//       return 0;
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        foo
        async () => {
          return 0;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   foo
					//   ;() => {
					//     return 0;
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        class A {
          foo() {}
          async [bar]() {
            baz;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   class A {
					//     foo() {}
					//     [bar]() {
					//       baz;
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
	})
}
