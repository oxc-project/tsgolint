# Contributing to tsgolint

Thank you for your interest in contributing to **tsgolint**! This document provides guidelines and information for contributors.

> [!IMPORTANT]
> **tsgolint** is a prototype in the early stages of development.
> This is a community effort. Feel free to ask to be assigned to any of the [good first issues](https://github.com/oxc-project/tsgolint/contribute).

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git with submodule support
- [just](https://github.com/casey/just) command runner
- Node.js and pnpm (for oxlint integration testing)
- Basic understanding of TypeScript and Go

### Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/oxc-project/tsgolint.git
   cd tsgolint
   ```

2. **Initialize and set up the project:**
   ```bash
   just init
   ```

3. **Build tsgolint:**
   ```bash
   just build
   ```

4. **Verify the build:**
   ```bash
   ./tsgolint --help
   ```

5. **Set up oxlint integration (optional, for integration testing):**
   ```bash
   # Install oxlint with type-aware support
   pnpm add -D oxlint-tsgolint@latest

   # Test integration with your locally built tsgolint
   OXLINT_TSGOLINT_PATH=./tsgolint pnpm dlx oxlint --type-aware

   # Or use the installed version
   pnpm dlx oxlint --type-aware
   ```

   > **Note:** Use `OXLINT_TSGOLINT_PATH` to test your local changes to tsgolint with oxlint.

## Development Workflow

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**

3. **Test your changes:**
   ```bash
   # Run all tests (unit + e2e)
   just test

   # Or run tests individually:
   go test ./internal/...  # Unit tests only

   # Test oxlint integration with your local tsgolint build
   OXLINT_TSGOLINT_PATH=./tsgolint oxlint --type-aware [test-files]
   ```

4. **Commit and push:**
   ```bash
   git add .
   git commit -m "feat: your feature description"
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request**

### Code Style

Use the provided just commands to ensure code quality:

```bash
# Format code
just fmt

# Run linter
just lint

# Run everything (format, lint, test)
just ready
```

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for complex logic
- Ensure all tests pass before submitting

## Testing

### Running Tests

```bash
# Run all tests (recommended)
just test

# Run only unit tests
go test ./internal/...

# Run tests with verbose output
go test -v ./internal/...
```

The test suite includes:

- Unit tests for individual rules
- End-to-end tests with real TypeScript code
- Integration tests with fixture files

### Testing Individual Rules

When developing or debugging a specific rule:

```bash
# Test a specific rule package
go test ./internal/rules/no_unsafe_argument

# Run with verbose output
go test -v ./internal/rules/no_unsafe_argument

# Run with debug logging
OXC_LOG=debug go test ./internal/rules/no_unsafe_argument
```

## Implementing New Rules

### Rule Structure

Each rule follows a consistent interface pattern:

```go
package rule_name

import (
    "github.com/oxc-project/tsgolint/internal/rule"
    "github.com/oxc-project/tsgolint/shim/ast"
)

type RuleName struct{}

func NewRuleName() *RuleName {
    return &RuleName{}
}

func (r *RuleName) Name() string {
    return "rule-name"
}

func (r *RuleName) Run(ctx rule.Context) rule.Listeners {
    return rule.Listeners{
        ast.NodeKind: func(node *ast.Node) {
            // Rule implementation
        },
    }
}
```

### Rule Development Guidelines

1. **Follow typescript-eslint compatibility:** Ensure behavior matches the corresponding typescript-eslint rule
2. **Use type-aware analysis:** Leverage the TypeScript checker for accurate type information
3. **Provide clear diagnostics:** Error messages should be helpful and actionable
4. **Support fixing:** When possible, implement automatic code fixes
5. **Add comprehensive tests:** Cover edge cases and different TypeScript constructs

### Adding a New Rule

1. **Create rule directory:**
   ```bash
   mkdir internal/rules/your_rule_name
   ```

2. **Implement the rule:**
   - Create `rule.go` with rule implementation
   - Create `rule_test.go` with comprehensive tests
   - Add test fixtures in appropriate directories

3. **Register the rule:**
   Add your rule to the rule registry in `cmd/tsgolint/main.go`

4. **Update documentation:**
   - Add rule to README.md rule list
   - Update rule count if implementing a new rule

### Test Fixtures

Create test fixtures that cover:

- Valid code that should not trigger the rule
- Invalid code that should trigger the rule
- Edge cases and complex TypeScript constructs
- Code that can be auto-fixed

## Debugging

### Debug Logging

Enable verbose debug output:

```bash
OXC_LOG=debug ./tsgolint [files...]
```

Debug logging provides information about:

- File assignment to TypeScript programs
- Worker distribution and execution
- Performance timing information
- Internal state details

### Common Issues

#### Build Issues

- **Patch application fails:** Ensure you're in the `typescript-go` directory when applying patches
- **Go build errors:** Check Go version (requires 1.21+) and ensure all dependencies are available

#### Test Issues

- **Integration tests fail:** Check that your changes don't break existing functionality
- **Unit tests timeout:** Some tests may take longer with debug logging enabled

#### Oxlint Integration Issues

- **oxlint --type-aware not working:** Ensure `oxlint-tsgolint` is installed and up to date
- **Testing local changes:** Use `OXLINT_TSGOLINT_PATH=./tsgolint` to point oxlint to your locally built binary
- **Performance issues with large repos:** This is a known limitation; consider testing with smaller codebases
- **Configuration not working:** Check `.oxlintrc.json` format and rule names

## TypeScript Integration

### Understanding the Shim Layer

**tsgolint** uses a shim layer to access internal typescript-go APIs. This is **not recommended for production use** but enables access to full TypeScript compiler functionality.

Key shim components:

- `shim/ast`: TypeScript AST node types and utilities
- `shim/checker`: Type checker bindings
- `shim/compiler`: Program and compilation host
- `shim/scanner`: Source text processing utilities

### Working with TypeScript Types

When implementing rules, you can access TypeScript type information:

```go
func (r *YourRule) checkNode(ctx rule.Context, node *ast.Node) {
    // Get type information
    nodeType := ctx.Checker.GetTypeAtLocation(node)

    // Check type properties
    if nodeType.IsString() {
        // Handle string types
    }

    // Get symbol information
    symbol := ctx.Checker.GetSymbolAtLocation(node)
    if symbol != nil {
        // Use symbol information
    }
}
```

## Performance Considerations

### Rule Performance

- Minimize expensive operations in hot paths
- Cache type checker results when possible
- Use appropriate AST node listeners (don't listen for all nodes if you only need specific ones)

### Testing Performance

Monitor rule performance:

```bash
# Profile CPU usage
go test -cpuprofile=cpu.prof ./internal/rules/your_rule

# Profile memory usage
go test -memprofile=mem.prof ./internal/rules/your_rule
```

## Communication

### Getting Help

- **Discord:** Join our [Discord server](https://discord.gg/9uXCAwqQZW) for real-time discussion
- **Issues:** Use GitHub issues for bug reports and feature requests
- **Discussions:** Use GitHub discussions for questions and general discussion

### Reporting Issues

When reporting issues, please include:

1. **Environment information:**
   - tsgolint version
   - Go version
   - Operating system
   - TypeScript version

2. **Reproduction steps:**
   - Minimal code example
   - Commands to reproduce
   - Expected vs actual behavior

3. **Debug output (if relevant):**
   ```bash
   OXC_LOG=debug ./tsgolint [files...]
   ```

## Pull Request Guidelines

### Before Submitting

- [ ] Run `just ready` to ensure everything passes
- [ ] Tests pass (`just test`)
- [ ] Code is formatted (`just fmt`)
- [ ] Linter passes (`just lint`)
- [ ] Changes are documented (if user-facing)
- [ ] Commit messages are clear and descriptive

### Pull Request Process

1. **Create descriptive title:** Use conventional commit format when possible
2. **Provide context:** Explain the problem and solution
3. **Link related issues:** Reference any related GitHub issues
4. **Request review:** Tag relevant maintainers or community members
5. **Address feedback:** Respond to review comments promptly

### Conventional Commits

We encourage using conventional commit format:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test-related changes
- `refactor:` for code refactoring
- `perf:` for performance improvements

Example: `feat: implement no-unnecessary-condition rule`

## Code of Conduct

Please note that this project follows the [Oxc Code of Conduct](https://github.com/oxc-project/oxc/blob/main/CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## License

By contributing to tsgolint, you agree that your contributions will be licensed under the MIT License.

## Additional Resources

- [ARCHITECTURE.md](./ARCHITECTURE.md) - Detailed technical documentation
- [typescript-go documentation](https://github.com/microsoft/typescript-go) - Understanding the underlying TypeScript compiler
- [typescript-eslint rules](https://typescript-eslint.io/rules/) - Reference for rule compatibility
- [Go testing](https://golang.org/pkg/testing/) - Go testing documentation
