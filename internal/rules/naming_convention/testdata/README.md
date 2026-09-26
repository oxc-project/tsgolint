# Upstream naming-convention cases

These 17 JSON suites contain every expanded `@typescript-eslint/naming-convention` RuleTester case from typescript-eslint commit `afb56de7d0123c79d2a2ffbb28ff7a8f8d951eab`: 8,966 valid cases, 7,146 invalid cases, and 44,065 expected diagnostics. The cases are checked in so Go tests do not need Node or an upstream checkout at test time. `manifest.json` records the source file hashes and a SHA-256 hash of each case's canonical JSON. `messages.json` contains the upstream message templates.

To regenerate from a checkout at that commit:

```sh
node --experimental-vm-modules tools/gen-naming-convention-tests.mjs /path/to/typescript-eslint
```

To audit the checked-in files against that checkout without writing files:

```sh
node --experimental-vm-modules tools/gen-naming-convention-tests.mjs /path/to/typescript-eslint --check
go test ./internal/rules/naming_convention -run '^TestGeneratedUpstreamCaseParity$' -count=1
```

The JSON keeps source code whitespace, options, parser options, diagnostic IDs, data, and positions. The Go loader maps upstream `languageOptions.parserOptions.project` to `tsconfig.naming-convention-project.json` (ES2015 target and ES2015/ES2017/ESNext libraries); cases without that setting use the ESNext fixture. An absent `options` field remains nil, an empty array remains an empty slice, and an explicit null remains distinguishable. Upstream errors with `data` assert the interpolated message; errors without `data` assert their message ID. The Go harness also checks reported locations for upstream errors that specify them.
