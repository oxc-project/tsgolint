# TSGoLint E2E Tests

Cross-platform end-to-end testing for TSGoLint using TypeScript and Vitest.

## Setup

```bash
cd e2e
pnpm install
```

## Running Tests

```bash
pnpm test
```

## How It Works

1. The test collects all TypeScript files from the `fixtures/` directory
2. Generates a configuration with all rules enabled for each file
3. Runs `tsgolint headless` with `GOMAXPROCS=1` for deterministic output
4. Parses the binary output to extract diagnostics
5. Sorts diagnostics deterministically for consistent snapshots
6. Compares the output with the expected snapshot

## Cross-Platform Compatibility

-   Uses Node.js built-in modules and cross-platform npm packages
-   Avoids shell-specific syntax (no bash required)
-   Works on Windows, macOS, and Linux
-   Path handling uses Node.js path module for OS-agnostic paths
