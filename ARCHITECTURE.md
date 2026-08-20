# Architecture

**tsgolint** is a high-performance TypeScript linter powered by the [native TypeScript compiler](https://github.com/microsoft/TypeScript/tree/main/tsc) and designed for integration with [Oxlint](https://oxc.rs/docs/guide/usage/linter.html).

## Overview

```
┌─────────────┐    ┌──────────────────────┐
│ Oxlint CLI  │◄──►│      tsgolint        │
│  (frontend) │    │     (backend)        │
│ • Files     │    │ • Type-aware rules   │
│ • Config    │    │ • Parallel workers   │
│ • Output    │    │ • TypeScript AST     │
└─────────────┘    └──────────────────────┘
```

**Frontend/Backend Architecture:**

- **Oxlint CLI (frontend)**: File discovery, configuration, output formatting, rule orchestration
- **tsgolint (backend)**: Type-aware rule execution, TypeScript integration, and diagnostic generation

This separation allows tsgolint to focus purely on type-aware analysis while Oxlint handles all the user-facing concerns like CLI, configuration, and output formatting.

## Core Architecture

### TypeScript Integration

tsgolint uses TypeScript's **native Go compiler** for native performance:

- **Direct AST**: No conversion overhead (TypeScript AST → rules)
- **Native Speed**: Go implementation with full TypeScript compiler
- **Type-aware**: Complete access to TypeScript type checker

### Parallel Processing

```
Coordinator → [Worker Pool] → Diagnostics
     ↓              ↓
Files + Rules → Rule Execution → Output
```

- **Worker Pool**: The CLI uses `runtime.GOMAXPROCS(0)` to set the worker count
- **Shared Programs**: Workers share a program and file queue, with a separate checker for each checker workload
- **Diagnostics**: Workers report diagnostics through callbacks; headless mode streams them to Oxlint

### Rule System

Rules follow a visitor pattern:

```go
var ExampleRule = rule.Rule{
	Name: "example-rule",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// Inspect the call using ctx.TypeChecker and report diagnostics.
			},
		}
	},
}
```

Each rule registers listeners for specific AST node types and uses the TypeScript checker for type-aware analysis.
The interfaces are defined in [`internal/rule/rule.go`](./internal/rule/rule.go); see [CONTRIBUTING.md](./CONTRIBUTING.md#implementing-new-rules) for a complete example.

## Key Design Decisions

### Why Go?

- **Performance**: Native compilation and direct access to native TypeScript; see the measured [benchmarks](./benchmarks/README.md)
- **Concurrency**: Excellent parallel processing primitives
- **Type Safety**: Compile-time checks for Go types and interfaces

### Why Direct TypeScript AST?

- **Zero Conversion**: No TypeScript → ESTree overhead
- **Complete Information**: Access to all TypeScript-specific data
- **Type Precision**: Better type-aware analysis

### Why Separate from Oxlint?

- **Clean Separation**: Independent development and testing
- **Focused Scope**: Type-aware rules and optional TypeScript diagnostics
- **Multiple Frontends**: Potential for other integrations

## TypeScript Shims

tsgolint accesses the native TypeScript compiler's internals via Go's `linkname` directives:

```
Go Shims → Native TypeScript Internal APIs → TypeScript Compiler
```

**Components:**

- `shim/ast`: TypeScript AST types
- `shim/checker`: Type checker interface
- `shim/compiler`: Program creation and management

The shims depend on internal APIs at the pinned native TypeScript revision. Regenerate them with `just shim` when their configuration or the upstream APIs change. See [tools/gen_shims/README.md](./tools/gen_shims/README.md).

Local native TypeScript adaptations are maintained in the [patch stack](./patches/README.md) and applied during `just init`.

## Performance Architecture

### Speed Sources

1. **Native Compilation**: Go → machine code
2. **Parallel Workers**: Multi-core utilization
3. **Zero Conversion**: Direct TypeScript AST usage
4. **Efficient Memory**: Streaming diagnostics

### Scalability

- **Work Distribution**: Files distributed across workers
- **Shared Programs**: TypeScript programs shared for efficiency
- **Memory Streaming**: Diagnostics processed immediately

## Maintenance Considerations

- **Version Synchronization**: Keep the native TypeScript revision, local patches, and generated shims in sync
- **Concurrency**: Keep mutable rule state local to each rule invocation and use the checker provided by `RuleContext`
- **Performance**: Profile changes to program creation, rule execution, and diagnostic reporting on representative projects

## References

- [TypeScript native compiler](https://github.com/microsoft/TypeScript/tree/main/tsc) - TypeScript compiler in Go
- [typescript-eslint](https://typescript-eslint.io/) - Rule compatibility reference
- [Oxlint](https://oxc.rs/) - Frontend CLI integration
