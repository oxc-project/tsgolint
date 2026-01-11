<div align="center">
  <h1>✨ tsgolint ✨</h1>
</div>

<div align="center">

[![MIT licensed][license-badge]][license-url]
[![Build Status][ci-badge]][ci-url]
[![Discord chat][discord-badge]][discord-url]
[![npm version][npm-badge]][npm-url]

</div>

**tsgolint** is a high-performance TypeScript linter containing only type-aware rules, powered by [typescript-go](https://github.com/microsoft/typescript-go) and designed for integration with [Oxlint](https://oxc.rs/docs/guide/usage/linter.html).

Key highlights:

- **Performance**: 20-40x faster than ESLint + typescript-eslint
- **Type-aware**: Comprehensive TypeScript type checking integration
- **Parallel**: Utilizes all available CPU cores
- **Compatible**: Implements typescript-eslint rules with consistent behavior

This project originated in [typescript-eslint/tsgolint](https://github.com/typescript-eslint/tsgolint). Fork permission is granted by @auvred.

> [!IMPORTANT]
> **tsgolint** is a prototype in the early stages of development.
> This is a community effort. Feel free to ask to be assigned to any of the [good first issues](https://github.com/oxc-project/tsgolint/contribute).

## Installation & Usage

**tsgolint** is integrated into Oxlint as the type-aware backend. Install and use via Oxlint:

```shell
# Install oxlint with type-aware support
pnpm add -D oxlint-tsgolint@latest

# Quick start
pnpm dlx oxlint --type-aware

# Or run on your project
oxlint --type-aware
```

### Configuration

Configure type-aware rules in `.oxlintrc.json`:

```json
{
  "$schema": "./node_modules/oxlint/configuration_schema.json",
  "rules": {
    "typescript/no-floating-promises": "error",
    "typescript/no-misused-promises": "error"
  }
}
```

Over 30 TypeScript-specific type-aware rules are available. For detailed setup and configuration, see the [Oxlint Type-Aware Linting guide](https://oxc.rs/blog/2025-08-17-oxlint-type-aware.html).

> [!NOTE]
> Non-type-aware TypeScript rules can be found in [Oxlint's TypeScript rules](https://oxc.rs/docs/guide/usage/linter/rules.html) under the TypeScript source.

## Performance

**tsgolint** is **20-40 times faster** than ESLint + typescript-eslint.

### Real-World Performance Examples

- **napi-rs** (144 files): 1.0s
- **preact** (245 files): 2.7s
- **rolldown** (314 files): 1.5s
- **bluesky** (1152 files): 7.0s

### Speed Sources

- **Native Speed**: Go implementation with direct TypeScript compiler integration
- **Zero Conversion**: Direct TypeScript AST usage without ESTree conversion overhead
- **Parallel Processing**: Multi-core execution utilizing all available CPU cores
- **Efficient Memory**: Streaming diagnostics and optimized resource usage

See [benchmarks](./benchmarks/README.md) for detailed performance comparisons.

## Current Status

### In Development 🚧

- Additional typescript-eslint rules
- Disable Comments
- Rule Configuration
- IDE usage

## Architecture

**tsgolint** follows a clean separation between frontend and backend:

- **Oxlint CLI** handles file discovery, configuration, and output formatting
- **tsgolint backend** provides type-aware rule execution and diagnostics
- **TypeScript integration** via typescript-go for native performance

For detailed technical documentation, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Contributing

We welcome contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for:

- Development setup and building instructions
- Testing procedures and guidelines
- How to implement new rules
- Code style and contribution workflow

## Implemented Rules

Implemented 45/59.

- [ ] [consistent-return](https://typescript-eslint.io/rules/consistent-return)
- [ ] [consistent-type-exports](https://typescript-eslint.io/rules/consistent-type-exports)
- [ ] [dot-notation](https://typescript-eslint.io/rules/dot-notation)
- [ ] [naming-convention](https://typescript-eslint.io/rules/naming-convention)
- [ ] [no-unnecessary-condition](https://typescript-eslint.io/rules/no-unnecessary-condition)
- [ ] [no-unnecessary-qualifier](https://typescript-eslint.io/rules/no-unnecessary-qualifier)
- [ ] [no-unnecessary-type-conversion](https://typescript-eslint.io/rules/no-unnecessary-type-conversion)
- [ ] [no-unnecessary-type-parameters](https://typescript-eslint.io/rules/no-unnecessary-type-parameters)
- [ ] [prefer-destructuring](https://typescript-eslint.io/rules/prefer-destructuring)
- [ ] [prefer-find](https://typescript-eslint.io/rules/prefer-find)
- [ ] [prefer-readonly](https://typescript-eslint.io/rules/prefer-readonly)
- [ ] [prefer-readonly-parameter-types](https://typescript-eslint.io/rules/prefer-readonly-parameter-types)
- [ ] [prefer-regexp-exec](https://typescript-eslint.io/rules/prefer-regexp-exec)
- [ ] [prefer-string-starts-ends-with](https://typescript-eslint.io/rules/prefer-string-starts-ends-with)
- [x] [await-thenable](https://typescript-eslint.io/rules/await-thenable)
- [x] [no-array-delete](https://typescript-eslint.io/rules/no-array-delete)
- [x] [no-base-to-string](https://typescript-eslint.io/rules/no-base-to-string)
- [x] [no-confusing-void-expression](https://typescript-eslint.io/rules/no-confusing-void-expression)
- [x] [no-deprecated](https://typescript-eslint.io/rules/no-deprecated)
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
- [x] [prefer-includes](https://typescript-eslint.io/rules/prefer-includes)
- [x] [prefer-nullish-coalescing](https://typescript-eslint.io/rules/prefer-nullish-coalescing)
- [x] [prefer-optional-chain](https://typescript-eslint.io/rules/prefer-optional-chain)
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
- [x] [strict-boolean-expressions](https://typescript-eslint.io/rules/strict-boolean-expressions)
- [x] [switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)
- [x] [unbound-method](https://typescript-eslint.io/rules/unbound-method)
- [x] [use-unknown-in-catch-callback-variable](https://typescript-eslint.io/rules/use-unknown-in-catch-callback-variable)

## Links

- [ARCHITECTURE.md](./ARCHITECTURE.md) - Detailed technical documentation
- [CONTRIBUTING.md](./CONTRIBUTING.md) - Development and contribution guidelines
- [Benchmarks](./benchmarks/README.md) - Performance comparison data
- [TypeScript Go](https://github.com/microsoft/typescript-go) - Underlying TypeScript compiler
- [Oxlint](https://oxc.rs/docs/guide/usage/linter.html) - Frontend linter integration

<!-- Badge definitions -->

[npm-badge]: https://img.shields.io/npm/v/oxlint-tsgolint.svg
[npm-url]: https://www.npmjs.com/package/oxlint-tsgolint
[license-badge]: https://img.shields.io/badge/license-MIT-blue.svg
[license-url]: LICENSE
[ci-badge]: https://github.com/oxc-project/tsgolint/actions/workflows/ci.yml/badge.svg
[ci-url]: https://github.com/oxc-project/tsgolint/actions/workflows/ci.yml
[discord-badge]: https://img.shields.io/discord/1079625926024900739?logo=discord&label=Discord
[discord-url]: https://discord.gg/9uXCAwqQZW
