# Architecture

**tsgolint** is a high-performance TypeScript linter powered by [typescript-go](https://github.com/microsoft/typescript-go) and designed for integration with [Oxlint](https://oxc.rs/docs/guide/usage/linter.html).

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

tsgolint uses **typescript-go** for native performance:

- **Direct AST**: No conversion overhead (TypeScript AST → rules)
- **Native Speed**: Go implementation with full TypeScript compiler
- **Type-aware**: Complete access to TypeScript type checker

### Parallel Processing

```
Master Thread → [Worker Pool] → Diagnostics
     ↓              ↓
Files + Rules → Rule Execution → Output
```

- **Worker Pool**: Utilizes all CPU cores
- **Shared State**: TypeScript programs shared across workers
- **Streaming**: Real-time diagnostic collection

### Rule System

Rules follow a visitor pattern:

```go
func (r *Rule) Run(ctx RuleContext) RuleListeners {
    return RuleListeners{
        ast.FunctionDeclaration: r.checkFunction,
        ast.CallExpression: r.checkCall,
    }
}
```

Each rule registers listeners for specific AST node types and uses the TypeScript checker for type-aware analysis.

## Key Design Decisions

### Why Go?

- **Performance**: 20-40x faster than JavaScript
- **Concurrency**: Excellent parallel processing primitives
- **Type Safety**: Prevents runtime errors

### Why Direct TypeScript AST?

- **Zero Conversion**: No TypeScript → ESTree overhead
- **Complete Information**: Access to all TypeScript-specific data
- **Type Precision**: Better type-aware analysis

### Why Separate from Oxlint?

- **Clean Separation**: Independent development and testing
- **Focused Scope**: tsgolint only handles type-aware rules
- **Multiple Frontends**: Potential for other integrations

## TypeScript Shims

tsgolint accesses typescript-go internals via Go's `linkname` directives:

```
Go Shims → typescript-go Internal APIs → TypeScript Compiler
```

**Components:**

- `shim/ast`: TypeScript AST types
- `shim/checker`: Type checker interface
- `shim/compiler`: Program creation and management

> **Note**: This approach is not recommended for production use. We're waiting for official typescript-go APIs.

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

## Known Limitations

- **Large Monorepos**: Performance issues with very large codebases
- **Memory Usage**: Potential issues with complex TypeScript configurations
- **Version Synchronization**: Must stay synchronized with TypeScript versions
- **Concurrency**: Rare potential for deadlocks in complex scenarios

## References

- [typescript-go](https://github.com/microsoft/typescript-go) - TypeScript compiler in Go
- [typescript-eslint](https://typescript-eslint.io/) - Rule compatibility reference
- [Oxlint](https://oxc.rs/) - Frontend CLI integration
