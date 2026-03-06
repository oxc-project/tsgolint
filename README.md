<div align="center">
  <h1>✨ tsgolint ✨</h1>
</div>

<div align="center">

[![MIT licensed][license-badge]][license-url]
[![Build Status][ci-badge]][ci-url]
[![Discord chat][discord-badge]][discord-url]
[![npm weekly downloads][npm-badge]][npm-url]

</div>

High-performance **type-aware linting** for Oxlint.

`tsgolint` executes lint rules that require **TypeScript semantic analysis**, using [typescript-go](https://github.com/microsoft/typescript-go) for full compatibility with the TypeScript type system, and targets **TypeScript 7** (codenamed **Project Corsa**).

It is designed to integrate seamlessly with Oxlint's fast syntax linting, enabling projects to run deeper semantic checks without sacrificing performance.

Key highlights:

- **Performance**: 20-40x faster than ESLint + typescript-eslint on large repositories
- **Coverage**: 59/61 targeted `typescript-eslint` type-aware rules implemented
- **Parallel**: Multi-core rule execution for scalable analysis
- **High impact**: catches production-grade bugs that syntax-only linting misses (for example `no-floating-promises`)

This project originated in [typescript-eslint/tsgolint](https://github.com/typescript-eslint/tsgolint), with fork permission granted by [@auvred](https://github.com/auvred).

## Why Teams Upgrade to Type-Aware Linting

If you ship TypeScript, running `oxlint --type-aware` in CI is a high-leverage upgrade that catches bug classes syntax linting cannot.

For example, `typescript/no-floating-promises` catches silently dropped async failures:

```js
async function saveUser(user) {
  const res = await fetch('/api/users', {
    method: 'POST',
    body: JSON.stringify(user),
  });
  if (!res.ok) throw new Error('save failed');
}

function onSubmit(user) {
  saveUser(user); // no-floating-promises: Promise is created but never handled
  showToast('Saved!'); // UI claims success even if the request rejects
}
```

Without this rule, rejected promises can be missed and reach production as flaky, hard-to-debug failures.

## Installation & Usage

`tsgolint` is integrated into Oxlint as the type-aware backend. Install and use via Oxlint:

```shell
# Install oxlint with type-aware support
pnpm add -D oxlint-tsgolint@latest

# Quick start
pnpm dlx oxlint --type-aware

# Or run on your project
oxlint --type-aware

# Optionally also run typechecking at the same time
oxlint --type-aware --type-check
```

### What these flags do

- `--type-aware`: enables `typescript/*` rules that require TypeScript semantic analysis via `tsgolint`
- `--type-check`: includes type diagnostics from `typescript-go` in type-aware runs

### Configuration

Configure type-aware rules in `.oxlintrc.json`:

```jsonc
{
  "$schema": "./node_modules/oxlint/configuration_schema.json",
  // alternatively, configure via options field
  "options": {
    "typeAware": true,
    "typeCheck": true,
  },
  "rules": {
    "typescript/no-floating-promises": "error",
    "typescript/no-misused-promises": "error",
  },
}
```

Over 50 TypeScript-specific type-aware rules are available. For detailed setup and configuration, see the [Oxlint Type-Aware Linting guide](https://oxc.rs/docs/guide/usage/linter/type-aware.html).

> [!NOTE]
> Non-type-aware TypeScript rules can be found in [Oxlint's TypeScript rules](https://oxc.rs/docs/guide/usage/linter/rules.html) under the TypeScript source.

## How it fits into Oxlint

Oxlint separates linting into two layers:

| Layer        | Purpose                      | Speed                  |
| ------------ | ---------------------------- | ---------------------- |
| **Oxlint**   | Syntax & structural analysis | Instant                |
| **tsgolint** | Type-aware semantic rules    | Requires type analysis |

This architecture keeps the common case extremely fast while enabling powerful type-aware checks when needed.
Oxlint handles file discovery, configuration, and output formatting, while `tsgolint` executes type-aware rules and emits semantic diagnostics.

## Why tsgolint exists

Traditional type-aware linting in the JavaScript ecosystem typically works by embedding TypeScript's type-analysis engine inside a JavaScript linter.

This approach introduces several bottlenecks:

- slow startup due to compiler initialization
- AST conversion between toolchains
- limited parallelism
- high memory overhead on large repositories

`tsgolint` takes a different approach: it runs directly on `typescript-go`, avoiding these bottlenecks and allowing semantic analysis to run efficiently alongside Oxlint.

Recent benchmark results (`eslint` + `typescript-eslint` vs `tsgolint`) show consistent large speedups:

| Repository           | ESLint + typescript-eslint | tsgolint | Speedup |
| -------------------- | -------------------------- | -------- | ------- |
| microsoft/vscode     | 167.8s                     | 4.89s    | **34x** |
| microsoft/typescript | 47.4s                      | 2.10s    | **23x** |
| typeorm/typeorm      | 27.3s                      | 0.93s    | **29x** |
| vuejs/core           | 20.7s                      | 0.95s    | **22x** |

See [benchmarks](./benchmarks/README.md) for detailed performance comparisons.

## Status

`tsgolint` is under active development.

The core architecture is stable and already powers Oxlint's type-aware linting mode. Current development focuses on expanding rule coverage and ecosystem integration:

- additional `typescript-eslint` rule support
- editor integration improvements
- configuration and rule customization
- performance improvements for large monorepos

Because `tsgolint` relies on `typescript-go`, its long-term stability evolves alongside TypeScript itself.

For detailed technical documentation, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Contributing

We welcome contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for:

- Development setup and building instructions
- Testing procedures and guidelines
- How to implement new rules
- Code style and contribution workflow

## Implemented Rules

Implemented 59/61.

- [ ] [naming-convention](https://typescript-eslint.io/rules/naming-convention)
- [ ] [prefer-destructuring](https://typescript-eslint.io/rules/prefer-destructuring)
- [x] [await-thenable](https://typescript-eslint.io/rules/await-thenable)
- [x] [consistent-return](https://typescript-eslint.io/rules/consistent-return)
- [x] [consistent-type-exports](https://typescript-eslint.io/rules/consistent-type-exports)
- [x] [dot-notation](https://typescript-eslint.io/rules/dot-notation)
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
- [x] [no-unnecessary-condition](https://typescript-eslint.io/rules/no-unnecessary-condition)
- [x] [no-unnecessary-qualifier](https://typescript-eslint.io/rules/no-unnecessary-qualifier)
- [x] [no-unnecessary-template-expression](https://typescript-eslint.io/rules/no-unnecessary-template-expression)
- [x] [no-unnecessary-type-arguments](https://typescript-eslint.io/rules/no-unnecessary-type-arguments)
- [x] [no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)
- [x] [no-unnecessary-type-conversion](https://typescript-eslint.io/rules/no-unnecessary-type-conversion)
- [x] [no-unnecessary-type-parameters](https://typescript-eslint.io/rules/no-unnecessary-type-parameters)
- [x] [no-unsafe-argument](https://typescript-eslint.io/rules/no-unsafe-argument)
- [x] [no-unsafe-assignment](https://typescript-eslint.io/rules/no-unsafe-assignment)
- [x] [no-unsafe-call](https://typescript-eslint.io/rules/no-unsafe-call)
- [x] [no-unsafe-enum-comparison](https://typescript-eslint.io/rules/no-unsafe-enum-comparison)
- [x] [no-unsafe-member-access](https://typescript-eslint.io/rules/no-unsafe-member-access)
- [x] [no-unsafe-return](https://typescript-eslint.io/rules/no-unsafe-return)
- [x] [no-unsafe-type-assertion](https://typescript-eslint.io/rules/no-unsafe-type-assertion)
- [x] [no-unsafe-unary-minus](https://typescript-eslint.io/rules/no-unsafe-unary-minus)
- [x] [no-useless-default-assignment](https://typescript-eslint.io/rules/no-useless-default-assignment)
- [x] [non-nullable-type-assertion-style](https://typescript-eslint.io/rules/non-nullable-type-assertion-style)
- [x] [only-throw-error](https://typescript-eslint.io/rules/only-throw-error)
- [x] [prefer-find](https://typescript-eslint.io/rules/prefer-find)
- [x] [prefer-includes](https://typescript-eslint.io/rules/prefer-includes)
- [x] [prefer-nullish-coalescing](https://typescript-eslint.io/rules/prefer-nullish-coalescing)
- [x] [prefer-optional-chain](https://typescript-eslint.io/rules/prefer-optional-chain)
- [x] [prefer-promise-reject-errors](https://typescript-eslint.io/rules/prefer-promise-reject-errors)
- [x] [prefer-readonly-parameter-types](https://typescript-eslint.io/rules/prefer-readonly-parameter-types)
- [x] [prefer-readonly](https://typescript-eslint.io/rules/prefer-readonly)
- [x] [prefer-reduce-type-parameter](https://typescript-eslint.io/rules/prefer-reduce-type-parameter)
- [x] [prefer-regexp-exec](https://typescript-eslint.io/rules/prefer-regexp-exec)
- [x] [prefer-return-this-type](https://typescript-eslint.io/rules/prefer-return-this-type)
- [x] [prefer-string-starts-ends-with](https://typescript-eslint.io/rules/prefer-string-starts-ends-with)
- [x] [promise-function-async](https://typescript-eslint.io/rules/promise-function-async)
- [x] [related-getter-setter-pairs](https://typescript-eslint.io/rules/related-getter-setter-pairs)
- [x] [require-array-sort-compare](https://typescript-eslint.io/rules/require-array-sort-compare)
- [x] [require-await](https://typescript-eslint.io/rules/require-await)
- [x] [restrict-plus-operands](https://typescript-eslint.io/rules/restrict-plus-operands)
- [x] [restrict-template-expressions](https://typescript-eslint.io/rules/restrict-template-expressions)
- [x] [return-await](https://typescript-eslint.io/rules/return-await)
- [x] [strict-boolean-expressions](https://typescript-eslint.io/rules/strict-boolean-expressions)
- [x] [strict-void-return](https://typescript-eslint.io/rules/strict-void-return)
- [x] [switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)
- [x] [unbound-method](https://typescript-eslint.io/rules/unbound-method)
- [x] [use-unknown-in-catch-callback-variable](https://typescript-eslint.io/rules/use-unknown-in-catch-callback-variable)

## Links

- [ARCHITECTURE.md](./ARCHITECTURE.md) - Detailed technical documentation
- [CONTRIBUTING.md](./CONTRIBUTING.md) - Development and contribution guidelines
- [Benchmarks](./benchmarks/README.md) - Performance comparison data
- [typescript-go](https://github.com/microsoft/typescript-go) - Underlying type-analysis backend
- [Oxlint](https://oxc.rs/docs/guide/usage/linter.html) - Frontend linter integration

<!-- Badge definitions -->

[npm-badge]: https://img.shields.io/npm/dw/oxlint-tsgolint.svg
[npm-url]: https://www.npmx.dev/package/oxlint-tsgolint
[license-badge]: https://img.shields.io/badge/license-MIT-blue.svg
[license-url]: LICENSE
[ci-badge]: https://github.com/oxc-project/tsgolint/actions/workflows/ci.yml/badge.svg
[ci-url]: https://github.com/oxc-project/tsgolint/actions/workflows/ci.yml
[discord-badge]: https://img.shields.io/discord/1079625926024900739?logo=discord&label=Discord
[discord-url]: https://discord.gg/9uXCAwqQZW
