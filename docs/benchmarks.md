# CodSpeed Benchmarks

This repository uses [CodSpeed](https://docs.codspeed.io/) for continuous benchmarking to track performance regressions over time.

## What is CodSpeed?

CodSpeed is a continuous benchmarking platform that automatically runs benchmarks on every pull request and commit, providing:

- Performance regression detection
- Historical performance tracking
- Detailed performance reports
- Integration with GitHub pull requests

## Benchmarks

The project includes Go benchmark tests for key performance-critical components:

### Core Linting (`internal/linter`)

- **BenchmarkLinterOnSingleFile**: Tests linting performance on a single TypeScript file
- **BenchmarkLinterParallel**: Tests parallel linting across multiple files
- **BenchmarkTypeChecking**: Tests TypeScript type checking operations
- **BenchmarkSourceCodeFixer**: Tests source code fixing functionality
- **BenchmarkSourceCodeFixerMultipleFixes**: Tests fixing with multiple fixes per diagnostic
- **BenchmarkSourceCodeFixerLargeFile**: Tests fixing performance on larger files

### Utilities (`internal/utils`)

- **BenchmarkCreateProgram**: Tests TypeScript program creation performance
- **BenchmarkCreateProgramMultipleFiles**: Tests program creation with multiple files
- **BenchmarkIsSymbolFromDefaultLibrary**: Tests symbol checking operations
- **BenchmarkASTTraversal**: Tests AST node processing operations  
- **BenchmarkOverlayVFS**: Tests virtual file system operations

## Running Benchmarks Locally

To run benchmarks locally:

```bash
# Run all benchmarks
go test -bench=. -run=^$ ./internal/utils ./internal/linter

# Run specific benchmark
go test -bench=BenchmarkLinterOnSingleFile -run=^$ ./internal/linter

# Run with memory allocation stats
go test -bench=. -benchmem -run=^$ ./internal/utils ./internal/linter
```

## CodSpeed Integration

The CodSpeed integration is configured in `.github/workflows/codspeed.yml` and runs automatically on:

- Every push to the `main` branch
- Every pull request

Performance results are automatically posted to pull requests and tracked over time in the CodSpeed dashboard.

## Performance Goals

tsgolint aims to be **20-40x faster** than ESLint + typescript-eslint through:

- Native Go implementation with TypeScript integration
- Parallel processing across all CPU cores
- Direct TypeScript AST usage (no AST conversion overhead)
- Efficient memory management and caching

The benchmarks help ensure these performance characteristics are maintained as the codebase evolves.