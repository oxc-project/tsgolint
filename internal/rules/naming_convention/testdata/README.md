# Upstream naming-convention cases

These 17 JSON suites preserve every `@typescript-eslint/naming-convention` RuleTester case from typescript-eslint commit `afb56de7d0123c79d2a2ffbb28ff7a8f8d951eab`: 8,966 valid cases, 7,146 invalid cases, and 44,065 expected diagnostics. The 16 generated suites store the upstream code templates and ordered options. A shared Go test helper expands them using the format names in `generator.json`, preserving every combination and duplicate. The main handwritten suite is stored directly. Go tests do not need Node or an upstream checkout.

`manifest.json` records source hashes, suite counts, and byte offsets into `fingerprints.bin`. That file contains one raw 32-byte SHA-256 digest per expanded case, in manifest suite order, with valid cases followed by invalid cases within each suite. The audit checks every generated Go case against its original canonical JSON fingerprint. Binary storage avoids thousands of lines of hexadecimal hashes without discarding any per-case verification. `messages.json` contains the upstream message templates.

To regenerate from a checkout at that commit:

```sh
node --experimental-vm-modules tools/gen-naming-convention-tests.mjs /path/to/typescript-eslint
```

To audit the checked-in files against that checkout without writing files:

```sh
node --experimental-vm-modules tools/gen-naming-convention-tests.mjs /path/to/typescript-eslint --check
go test ./internal/rules/naming_convention -run '^TestGeneratedUpstreamCaseParity$' -count=1
```

Expansion preserves source code whitespace, ordered options, parser options, diagnostic IDs, data, and positions. The Go loader maps upstream `languageOptions.parserOptions.project` to `tsconfig.naming-convention-project.json` (ES2015 target and ES2015/ES2017/ESNext libraries); cases without that setting use the ESNext fixture. An absent `options` field remains nil, an empty array remains an empty slice, and an explicit null remains distinguishable. Upstream errors with `data` assert the interpolated message; errors without `data` assert their message ID. The Go harness also checks reported locations for upstream errors that specify them. These explicit upstream assertions replace rendered diagnostic snapshots for this corpus; local regression snapshots remain separate.
