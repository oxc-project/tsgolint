# ✨ tsgolint ✨

**tsgolint** is a TypeScript linter containing only type-ware rules, powered by [typescript-go](https://github.com/microsoft/typescript-go) adapted for [Oxlint](https://oxc.rs/docs/guide/usage/linter.html).

This project originated in [typescript-eslint/tsgolint](https://github.com/typescript-eslint/tsgolint). Fork permission is granted by @auvred.

> If you want faster typed linting with ESLint, see [typescript-eslint/typescript-eslint#10940 Enhancement: Use TypeScript's Go port (tsgo / typescript-go) for type information](https://github.com/typescript-eslint/typescript-eslint/issues/10940).

> [!IMPORTANT]
> **tsgolint** is a prototype in the early stages of development.
> This is a community effort. Feel free to ask for assignee in any of the [good first issues](https://github.com/oxc-project/tsgolint/contribute).

![Running tsgolint on microsoft/typescript repo](./docs/record.gif)

## What's been prototyped

- Primitive linter engine
- Lint rules tester
- Source code fixer
- 40 [type-aware](https://typescript-eslint.io/blog/typed-linting) typescript-eslint's rules
- Basic `tsgolint` CLI

### Speedup over ESLint

**tsgolint** is **20-40 times faster** than ESLint + typescript-eslint.

Most of the speedup is due to the following facts:

- Native speed parsing and type-checking (thanks to [typescript-go](https://github.com/microsoft/typescript-go))
- No more [TS AST -> ESTree AST](https://typescript-eslint.io/blog/asts-and-typescript-eslint/#ast-formats) conversions. TS AST is directly used in rules.
- Parallel parsing, type checking and linting. **tsgolint** uses all available CPU cores.

See [benchmarks](./benchmarks/README.md) for more info.

## What hasn't been prototyped

- Non-type-aware rules
- Editor extension
- Rich CLI features
- Config file
- Plugin system

## Architecture

Oxlint CLI -> paths + rules -> tsgolint -> diagnostics -> Oxlint CLI.

* Oxlint is the "frontend" for tsgolint, it handles cli, paths walk, ignore logic and printing of diagnostics.
* tsgolint is the backend for Oxlint, accepting paths as input, and output structured diagnostics.

Scope of tsgolint is only:

* run the type aware rules
* pass structured diagnostics back to oxlint

## Note about typescript-go shims

tsgolint accesses internal typescript-go APIs via [shims](./tools/gen_shims/main.go), which is not a recommended approach.
We advise against using this method and suggest waiting for an official API release instead.
See [this discussion](https://github.com/microsoft/typescript-go/discussions/455) and [this one](https://github.com/microsoft/typescript-go/discussions/481).

More technical details can be found [here](./tools/gen_shims/README.md).

## Building `tsgolint`

```bash
git submodule update --init                       # init typescript-go submodule

cd typescript-go
git am --3way --no-gpg-sign ../patches/*.patch    # apply typescript-go patches
cd ..

go build -o tsgolint ./cmd/tsgolint
```

## Implemented rules

| Name                                                                                                                | Status |
| ------------------------------------------------------------------------------------------------------------------- | ------ |
| [await-thenable](https://typescript-eslint.io/rules/await-thenable)                                                 | ✅     |
| [no-array-delete](https://typescript-eslint.io/rules/no-array-delete)                                               | ✅     |
| [no-base-to-string](https://typescript-eslint.io/rules/no-base-to-string)                                           | ✅     |
| [no-confusing-void-expression](https://typescript-eslint.io/rules/no-confusing-void-expression)                     | ✅     |
| [no-duplicate-type-constituents](https://typescript-eslint.io/rules/no-duplicate-type-constituents)                 | ✅     |
| [no-floating-promises](https://typescript-eslint.io/rules/no-floating-promises)                                     | ✅     |
| [no-for-in-array](https://typescript-eslint.io/rules/no-for-in-array)                                               | ✅     |
| [no-implied-eval](https://typescript-eslint.io/rules/no-implied-eval)                                               | ✅     |
| [no-meaningless-void-operator](https://typescript-eslint.io/rules/no-meaningless-void-operator)                     | ✅     |
| [no-misused-promises](https://typescript-eslint.io/rules/no-misused-promises)                                       | ✅     |
| [no-misused-spread](https://typescript-eslint.io/rules/no-misused-spread)                                           | ✅     |
| [no-mixed-enums](https://typescript-eslint.io/rules/no-mixed-enums)                                                 | ✅     |
| [no-redundant-type-constituents](https://typescript-eslint.io/rules/no-redundant-type-constituents)                 | ✅     |
| [no-unnecessary-boolean-literal-compare](https://typescript-eslint.io/rules/no-unnecessary-boolean-literal-compare) | ✅     |
| [no-unnecessary-template-expression](https://typescript-eslint.io/rules/no-unnecessary-template-expression)         | ✅     |
| [no-unnecessary-type-arguments](https://typescript-eslint.io/rules/no-unnecessary-type-arguments)                   | ✅     |
| [no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)                   | ✅     |
| [no-unsafe-argument](https://typescript-eslint.io/rules/no-unsafe-argument)                                         | ✅     |
| [no-unsafe-assignment](https://typescript-eslint.io/rules/no-unsafe-assignment)                                     | ✅     |
| [no-unsafe-call](https://typescript-eslint.io/rules/no-unsafe-call)                                                 | ✅     |
| [no-unsafe-enum-comparison](https://typescript-eslint.io/rules/no-unsafe-enum-comparison)                           | ✅     |
| [no-unsafe-member-access](https://typescript-eslint.io/rules/no-unsafe-member-access)                               | ✅     |
| [no-unsafe-return](https://typescript-eslint.io/rules/no-unsafe-return)                                             | ✅     |
| [no-unsafe-type-assertion](https://typescript-eslint.io/rules/no-unsafe-type-assertion)                             | ✅     |
| [no-unsafe-unary-minus](https://typescript-eslint.io/rules/no-unsafe-unary-minus)                                   | ✅     |
| [non-nullable-type-assertion-style](https://typescript-eslint.io/rules/non-nullable-type-assertion-style)           | ✅     |
| [only-throw-error](https://typescript-eslint.io/rules/only-throw-error)                                             | ✅     |
| [prefer-promise-reject-errors](https://typescript-eslint.io/rules/prefer-promise-reject-errors)                     | ✅     |
| [prefer-reduce-type-parameter](https://typescript-eslint.io/rules/prefer-reduce-type-parameter)                     | ✅     |
| [prefer-return-this-type](https://typescript-eslint.io/rules/prefer-return-this-type)                               | ✅     |
| [promise-function-async](https://typescript-eslint.io/rules/promise-function-async)                                 | ✅     |
| [related-getter-setter-pairs](https://typescript-eslint.io/rules/related-getter-setter-pairs)                       | ✅     |
| [require-array-sort-compare](https://typescript-eslint.io/rules/require-array-sort-compare)                         | ✅     |
| [require-await](https://typescript-eslint.io/rules/require-await)                                                   | ✅     |
| [restrict-plus-operands](https://typescript-eslint.io/rules/restrict-plus-operands)                                 | ✅     |
| [restrict-template-expressions](https://typescript-eslint.io/rules/restrict-template-expressions)                   | ✅     |
| [return-await](https://typescript-eslint.io/rules/return-await)                                                     | ✅     |
| [switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)                       | ✅     |
| [unbound-method](https://typescript-eslint.io/rules/unbound-method)                                                 | ✅     |
| [use-unknown-in-catch-callback-variable](https://typescript-eslint.io/rules/use-unknown-in-catch-callback-variable) | ✅     |
