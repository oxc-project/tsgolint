package no_mixed_enums

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoMixedEnumsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoMixedEnumsRule, []rule_tester.ValidTestCase{
		{Code: `
      enum Fruit {}
    `},
		{Code: `
      enum Fruit {
        Apple,
      }
    `},
		{Code: `
      enum Fruit {
        Apple = false,
      }
    `},
		{Code: `
      enum Fruit {
        Apple,
        Banana,
      }
    `},
		{Code: `
      enum Fruit {
        Apple = 0,
        Banana,
      }
    `},
		{Code: `
      enum Fruit {
        Apple,
        Banana = 1,
      }
    `},
		{Code: `
      enum Fruit {
        Apple = 0,
        Banana = 1,
      }
    `},
		{Code: `
      enum Fruit {
        Apple,
        Banana = false,
      }
    `},
		{Code: `
      const getValue = () => 0;
      enum Fruit {
        Apple,
        Banana = getValue(),
      }
    `},
		{Code: `
      const getValue = () => 0;
      enum Fruit {
        Apple = getValue(),
        Banana = getValue(),
      }
    `},
		{Code: `
      const getValue = () => '';
      enum Fruit {
        Apple = '',
        Banana = getValue(),
      }
    `},
		{Code: `
      const getValue = () => '';
      enum Fruit {
        Apple = getValue(),
        Banana = '',
      }
    `},
		{Code: `
      const getValue = () => '';
      enum Fruit {
        Apple = getValue(),
        Banana = getValue(),
      }
    `},
		{Code: `
      enum First {
        A = 1,
      }

      enum Second {
        A = First.A,
        B = 2,
      }
    `},
		{Code: `
      enum First {
        A = '',
      }

      enum Second {
        A = First.A,
        B = 'b',
      }
    `},
		{Code: `
      enum Foo {
        A,
      }
      enum Foo {
        B,
      }
    `},
		{Code: `
      enum Foo {
        A = 0,
      }
      enum Foo {
        B,
      }
    `},
		{Code: `
      enum Foo {
        A,
      }
      enum Foo {
        B = 1,
      }
    `},
		{Code: `
      enum Foo {
        A = 0,
      }
      enum Foo {
        B = 1,
      }
    `},
		{Code: `
      enum Foo {
        A = 'a',
      }
      enum Foo {
        B = 'b',
      }
    `},
		{Code: `
      declare const Foo: any;
      enum Foo {
        A,
      }
    `},
		{Code: `
enum Foo {
  A = 1,
}
enum Foo {
  B = 2,
}
    `},
		{Code: `
enum Foo {
  A = ` + "`" + `A` + "`" + `,
}
enum Foo {
  B = ` + "`" + `B` + "`" + `,
}
    `},
		{Code: `
enum Foo {
  A = false, // (TS error)
}
enum Foo {
  B = ` + "`" + `B` + "`" + `,
}
    `},
		{Code: `
enum Foo {
  A = 'A',
}
enum Foo {
  B = false, // (TS error)
}
    `},
		{Code: `
import { Enum } from './mixed-enums-decl';

declare module './mixed-enums-decl' {
  enum Enum {
    StringLike = 'StringLike',
  }
}
    `},
		{Code: `
import { Enum } from "module-that-does't-exist";

declare module "module-that-doesn't-exist" {
  enum Enum {
    StringLike = 'StringLike',
  }
}
    `},
		{Code: `
namespace Test {
  export enum Bar {
    A = 1,
  }
}
namespace Test {
  export enum Bar {
    B = 2,
  }
}
    `},
		{Code: `
namespace Outer {
  namespace Test {
    export enum Bar {
      A = 1,
    }
  }
}
namespace Outer {
  namespace Test {
    export enum Bar {
      B = 'B',
    }
  }
}
    `},
		{Code: `
namespace Outer {
  namespace Test {
    export enum Bar {
      A = 1,
    }
  }
}
namespace Different {
  namespace Test {
    export enum Bar {
      B = 'B',
    }
  }
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        enum Fruit {
          Apple,
          Banana = 'banana',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      4,
					Column:    20,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
          Banana = 'banana',
          Cherry = 'cherry',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      4,
					Column:    20,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple,
          Banana,
          Cherry = 'cherry',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      5,
					Column:    20,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
        enum Fruit {
          Apple = 0,
          Banana = 'banana',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      4,
					Column:    20,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
        const getValue = () => 0;
        enum Fruit {
          Apple = getValue(),
          Banana = 'banana',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      5,
					Column:    20,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
        const getValue = () => '';
        enum Fruit {
          Apple,
          Banana = getValue(),
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      5,
					Column:    20,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
        const getValue = () => '';
        enum Fruit {
          Apple = getValue(),
          Banana = 0,
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      5,
					Column:    20,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
        enum First {
          A = 1,
        }

        enum Second {
          A = First.A,
          B = 'b',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      8,
					Column:    15,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
        enum First {
          A = 'a',
        }

        enum Second {
          A = First.A,
          B = 1,
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      8,
					Column:    15,
					EndColumn: 16,
				},
			},
		},
		{
			Code: `
        enum Foo {
          A,
        }
        enum Foo {
          B = 'b',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      6,
					Column:    15,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
        enum Foo {
          A = 1,
        }
        enum Foo {
          B = 'b',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      6,
					Column:    15,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
        enum Foo {
          A = 'a',
        }
        enum Foo {
          B,
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      6,
					Column:    11,
					EndColumn: 12,
				},
			},
		},
		{
			Code: `
        enum Foo {
          A = 'a',
        }
        enum Foo {
          B = 0,
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      6,
					Column:    15,
					EndColumn: 16,
				},
			},
		},
		{
			Code: `
        enum Foo {
          A,
        }
        enum Foo {
          B = 'b',
        }
        enum Foo {
          C = 'c',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      6,
					Column:    15,
					EndColumn: 18,
				},
				{
					MessageId: "mixed",
					Line:      9,
					Column:    15,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
        enum Foo {
          A,
        }
        enum Foo {
          B = 'b',
        }
        enum Foo {
          C,
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      6,
					Column:    15,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
        enum Foo {
          A,
        }
        enum Foo {
          B,
        }
        enum Foo {
          C = 'c',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      9,
					Column:    15,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
import { Enum } from './mixed-enums-decl';

declare module './mixed-enums-decl' {
  enum Enum {
    Numeric = 0,
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      6,
					Column:    15,
					EndColumn: 16,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1,
}
enum Foo {
  B = 'B',
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      6,
					Column:    7,
					EndColumn: 10,
				},
			},
		},
		{
			Code: `
namespace Test {
  export enum Bar {
    A = 1,
  }
}
namespace Test {
  export enum Bar {
    B = 'B',
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      9,
					Column:    9,
					EndColumn: 12,
				},
			},
		},
		{
			Code: `
namespace Test {
  export enum Bar {
    A,
  }
}
namespace Test {
  export enum Bar {
    B = 'B',
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      9,
					Column:    9,
					EndColumn: 12,
				},
			},
		},
		{
			Code: `
namespace Outer {
  export namespace Test {
    export enum Bar {
      A = 1,
    }
  }
}
namespace Outer {
  export namespace Test {
    export enum Bar {
      B = 'B',
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mixed",
					Line:      12,
					Column:    11,
					EndColumn: 14,
				},
			},
		},
	})
}
