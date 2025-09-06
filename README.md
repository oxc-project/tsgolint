# ✨ tsgolint ✨

**tsgolint** is a TypeScript linter containing only type-aware rules, powered by [typescript-go](https://github.com/microsoft/typescript-go) adapted for [Oxlint](https://oxc.rs/docs/guide/usage/linter.html).

This project originated in [typescript-eslint/tsgolint](https://github.com/typescript-eslint/tsgolint). Fork permission is granted by @auvred.

> If you want faster typed linting with ESLint, see [typescript-eslint/typescript-eslint#10940 Enhancement: Use TypeScript's Go port (tsgo / typescript-go) for type information](https://github.com/typescript-eslint/typescript-eslint/issues/10940).

> [!IMPORTANT]
> **tsgolint** is a prototype in the early stages of development.
> This is a community effort. Feel free to ask to be assigned to any of the [good first issues](https://github.com/oxc-project/tsgolint/contribute).

![Running tsgolint on microsoft/typescript repo](./docs/record.gif)

## What's been prototyped

- Primitive linter engine
- Lint rules tester
- Source code fixer
- 40 [type-aware](https://typescript-eslint.io/blog/typed-linting) typescript-eslint's rules
- Basic `tsgolint` CLI

Try running

```shell
npx oxlint-tsgolint --help
```

to see available options.

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

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architectural documentation.

## Building `tsgolint`

```bash
git submodule update --init                       # init typescript-go submodule

cd typescript-go
git am --3way --no-gpg-sign ../patches/*.patch    # apply typescript-go patches
cd ..

go build -o tsgolint ./cmd/tsgolint
```

## Debugging

For troubleshooting and development purposes, **tsgolint** supports debug logging through the `OXC_LOG` environment variable.

### Debug Logging

To enable verbose debug output, set the `OXC_LOG` environment variable to `debug`:

```bash
OXC_LOG=debug tsgolint
```

Debug logging provides detailed information about the linting process, including:

- File assignment to TypeScript programs
- Worker distribution and execution
- Performance timing information
- Internal state details

This can be helpful when:

- Diagnosing performance issues
- Understanding how files are being processed
- Troubleshooting TypeScript configuration problems
- Contributing to **tsgolint** development

## Testing

**tsgolint** includes several types of tests to ensure correctness:

### Unit Tests

Run Go unit tests for individual rules:

```shell
go test ./internal/...
```

### Integration Tests

- `./test.sh` - End-to-end snapshot test running all 40+ rules
  - Runs **all** tsgolint rules on all fixture files
  - Captures diagnostic output in deterministic, sortable format
  - Verifies output matches expected snapshot

## Implemented rules (40/59)

- [ ] [consistent-return](https://typescript-eslint.io/rules/consistent-return)
- [ ] [consistent-type-exports](https://typescript-eslint.io/rules/consistent-type-exports)
- [ ] [dot-notation](https://typescript-eslint.io/rules/dot-notation)
- [ ] [naming-convention](https://typescript-eslint.io/rules/naming-convention)
- [ ] [no-deprecated](https://typescript-eslint.io/rules/no-deprecated)
- [ ] [no-unnecessary-condition](https://typescript-eslint.io/rules/no-unnecessary-condition)
- [ ] [no-unnecessary-qualifier](https://typescript-eslint.io/rules/no-unnecessary-qualifier)
- [ ] [no-unnecessary-type-conversion](https://typescript-eslint.io/rules/no-unnecessary-type-conversion)
- [ ] [no-unnecessary-type-parameters](https://typescript-eslint.io/rules/no-unnecessary-type-parameters)
- [ ] [prefer-destructuring](https://typescript-eslint.io/rules/prefer-destructuring)
- [ ] [prefer-find](https://typescript-eslint.io/rules/prefer-find)
- [ ] [prefer-includes](https://typescript-eslint.io/rules/prefer-includes)
- [ ] [prefer-nullish-coalescing](https://typescript-eslint.io/rules/prefer-nullish-coalescing)
- [ ] [prefer-optional-chain](https://typescript-eslint.io/rules/prefer-optional-chain)
- [ ] [prefer-readonly](https://typescript-eslint.io/rules/prefer-readonly)
- [ ] [prefer-readonly-parameter-types](https://typescript-eslint.io/rules/prefer-readonly-parameter-types)
- [ ] [prefer-regexp-exec](https://typescript-eslint.io/rules/prefer-regexp-exec)
- [ ] [prefer-string-starts-ends-with](https://typescript-eslint.io/rules/prefer-string-starts-ends-with)
- [ ] [strict-boolean-expressions](https://typescript-eslint.io/rules/strict-boolean-expressions)
- [x] [await-thenable](https://typescript-eslint.io/rules/await-thenable)
- [x] [no-array-delete](https://typescript-eslint.io/rules/no-array-delete)
- [x] [no-base-to-string](https://typescript-eslint.io/rules/no-base-to-string)
- [x] [no-confusing-void-expression](https://typescript-eslint.io/rules/no-confusing-void-expression)
- [x] [no-duplicate-type-constituents](https://typescript-eslint.io/rules/no-duplicate-type-constituents)
- [x] [no-floating-promises](https://typescript-eslint.io/rules/no-floating-promises)
- [x] [no-for-in-array](https://typescript-eslint.io/rules/no-for-in-array)
- [x] [no-implied-eval](https://typescript-eslint.io/rules/no-implied-eval)
- [x] [no-meaningless-void-operator](https://typescript-eslint.io/rules/no-meaningless-void-operator)
- [x] [no-misused-promises](https://typescript-eslint.io/rules/no-misused-promises)
- [x] [no-misused-spread](https://typescript-eslint.io/rules/no-misused-spread)
- [x] [no-mixed-enums](https://typescript-eslint.io/rules/no-mixed-enums)
- [x] [no-redundant-type-constituents](https://typescript-eslint.io/rules/no-redundant-type-constituents)
- [x] [no-unnecessary-boolean-literal-compare](https://typescript-eslint.io/rules/no-unnecessary-boolean-literal-compare)
- [x] [no-unnecessary-template-expression](https://typescript-eslint.io/rules/no-unnecessary-template-expression)
- [x] [no-unnecessary-type-arguments](https://typescript-eslint.io/rules/no-unnecessary-type-arguments)
- [x] [no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)
- [x] [no-unsafe-argument](https://typescript-eslint.io/rules/no-unsafe-argument)
- [x] [no-unsafe-assignment](https://typescript-eslint.io/rules/no-unsafe-assignment)
- [x] [no-unsafe-call](https://typescript-eslint.io/rules/no-unsafe-call)
- [x] [no-unsafe-enum-comparison](https://typescript-eslint.io/rules/no-unsafe-enum-comparison)
- [x] [no-unsafe-member-access](https://typescript-eslint.io/rules/no-unsafe-member-access)
- [x] [no-unsafe-return](https://typescript-eslint.io/rules/no-unsafe-return)
- [x] [no-unsafe-type-assertion](https://typescript-eslint.io/rules/no-unsafe-type-assertion)
- [x] [no-unsafe-unary-minus](https://typescript-eslint.io/rules/no-unsafe-unary-minus)
- [x] [non-nullable-type-assertion-style](https://typescript-eslint.io/rules/non-nullable-type-assertion-style)
- [x] [only-throw-error](https://typescript-eslint.io/rules/only-throw-error)
- [x] [prefer-promise-reject-errors](https://typescript-eslint.io/rules/prefer-promise-reject-errors)
- [x] [prefer-reduce-type-parameter](https://typescript-eslint.io/rules/prefer-reduce-type-parameter)
- [x] [prefer-return-this-type](https://typescript-eslint.io/rules/prefer-return-this-type)
- [x] [promise-function-async](https://typescript-eslint.io/rules/promise-function-async)
- [x] [related-getter-setter-pairs](https://typescript-eslint.io/rules/related-getter-setter-pairs)
- [x] [require-array-sort-compare](https://typescript-eslint.io/rules/require-array-sort-compare)
- [x] [require-await](https://typescript-eslint.io/rules/require-await)
- [x] [restrict-plus-operands](https://typescript-eslint.io/rules/restrict-plus-operands)
- [x] [restrict-template-expressions](https://typescript-eslint.io/rules/restrict-template-expressions)
- [x] [return-await](https://typescript-eslint.io/rules/return-await)
- [x] [switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)
- [x] [unbound-method](https://typescript-eslint.io/rules/unbound-method)
- [x] [use-unknown-in-catch-callback-variable](https://typescript-eslint.io/rules/use-unknown-in-catch-callback-variable)
