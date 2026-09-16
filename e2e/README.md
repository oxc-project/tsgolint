# tsgolint End-to-End Tests

Cross-platform end-to-end testing for tsgolint using TypeScript and Vitest.

## Setup

From the repository root, run `just init` once to initialize the project and install dependencies, then `just build` to create the binary. The tests use `tsgolint` (or `tsgolint.exe` on Windows) from the repository root. Rebuild it after changing Go code.

To install only the end-to-end dependencies, run `pnpm --dir e2e --ignore-workspace install`.

## Running Tests

```bash
cd e2e

# Run once
pnpm test --run

# Watch for changes
pnpm test

# Update Vitest snapshots
pnpm test --run --update
```

From the repository root, `just test` rebuilds the binary and runs both the end-to-end and Go tests. `just update-snaps` updates only the Go test snapshots.

## How It Works

The main snapshot test in [`snapshot.test.ts`](./snapshot.test.ts):

1. Collects TypeScript files from `fixtures/basic/`
2. Generates a headless payload with all rules enabled
3. Runs `tsgolint headless -fix -fix-suggestions` with `GOMAXPROCS=1`
4. Parses the framed output to extract JSON diagnostics
5. Sorts diagnostics for consistent snapshots
6. Compares the output with the expected snapshot

Other tests cover rule options, project references, source overrides, TypeScript diagnostics, and regressions. When adding a rule, update `ALL_RULES` and add a fixture under `fixtures/basic/rules/`.

## Cross-Platform Compatibility

- Uses Node.js built-in modules and cross-platform npm packages
- Avoids shell-specific syntax (no bash required)
- Works on Windows, macOS, and Linux
- Path handling uses Node.js path module for OS-agnostic paths
