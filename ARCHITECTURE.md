# Architecture

**tsgolint** is a high-performance TypeScript linter containing only type-aware rules, powered by [typescript-go](https://github.com/microsoft/typescript-go) and designed for integration with [Oxlint](https://oxc.rs/docs/guide/usage/linter.html).

## tldr;

Oxlint CLI -> paths + rules -> tsgolint -> diagnostics -> Oxlint CLI.

* Oxlint is the "frontend" for tsgolint, it handles CLI, path walking, ignores logic and printing of diagnostics.
* tsgolint is the backend for Oxlint, accepting paths as input, and outputs structured diagnostics.

Scope of tsgolint is only:

* run the type-aware rules
* pass structured diagnostics back to oxlint

### Shimming typescript-go

`tsgolint` accesses internal typescript-go APIs via [shims](./tools/gen_shims/main.go), which is not a recommended approach.
We advise against using this method and suggest waiting for an official API release instead.
See [this discussion](https://github.com/microsoft/typescript-go/discussions/455) and [this one](https://github.com/microsoft/typescript-go/discussions/481).

More technical details can be found [here](./tools/gen_shims/README.md).

## Context & Goals

### Problem Statement
Traditional TypeScript linting with ESLint + typescript-eslint suffers from significant performance bottlenecks:
- AST conversion overhead (TypeScript AST → ESTree AST)
- Single-threaded execution model
- JavaScript runtime limitations

### Goals
- **Performance**: Achieve 20-40x speedup over ESLint + typescript-eslint
- **Compatibility**: Maintain compatibility with typescript-eslint rules
- **Integration**: Seamless backend integration with Oxlint frontend
- **Type Safety**: Leverage native TypeScript type checking capabilities

## Quality Attributes

### Performance
- **Parallel Processing**: Utilizes all available CPU cores
- **Native Speed**: Go implementation with direct TypeScript compiler integration
- **Zero Conversion**: Direct use of TypeScript AST without conversion overhead
- **Memory Efficiency**: Streaming diagnostics and buffered output

### Maintainability
- **Clear Separation**: Distinct CLI frontend and linting backend
- **Modular Rules**: Individual rule implementations with consistent interfaces
- **Type Safety**: Go's type system prevents many runtime errors

## System Overview

```
┌─────────────┐    ┌──────────────────────────────────────────────┐
│ Oxlint CLI  │    │                tsgolint                      │
│             │    │                                              │
│ • File      │◄──►│  ┌─────────────┐  ┌─────────────────────────┐ │
│   Discovery │    │  │   CLI       │  │     Linting Engine      │ │
│ • Ignores   │    │  │             │  │                         │ │
│ • Output    │    │  │ • Args      │  │ • Worker Pool           │ │
│   Formatting│    │  │ • Config    │──┤ • Rule Orchestration   │ │
│             │    │  │ • Files     │  │ • Parallel Processing   │ │
└─────────────┘    │  └─────────────┘  └─────────────────────────┘ │
                   │           │                      │              │
                   │           ▼                      ▼              │
                   │  ┌─────────────┐  ┌─────────────────────────┐ │
                   │  │ TypeScript  │  │      Rule System        │ │
                   │  │ Integration │  │                         │ │
                   │  │             │  │ • Rule Interface        │ │
                   │  │ • AST       │◄─┤ • Visitor Pattern      │ │
                   │  │ • Checker   │  │ • Context Management    │ │
                   │  │ • Program   │  │ • 40+ Type-aware Rules  │ │
                   │  └─────────────┘  └─────────────────────────┘ │
                   └──────────────────────────────────────────────┘
```

## Detailed Component Design

### 1. CLI Layer (`cmd/tsgolint`)
**Responsibilities:**
- Command-line argument parsing and validation
- TypeScript configuration resolution (`tsconfig.json`)
- File discovery and filtering (excludes `node_modules`)
- Performance profiling and tracing
- Output formatting and error reporting

**Key Files:**
- `main.go`: Entry point, CLI logic, rule registration
- `service.go`: Headless service mode for integration
- `headless.go`: API for programmatic usage

### 2. Linting Engine (`internal/linter`)
**Responsibilities:**
- Coordinating parallel execution across multiple workers
- Managing TypeScript program and checker instances
- Distributing file processing workload
- Collecting and streaming diagnostics

**Architecture Pattern:** Worker Pool
```go
type Workload = map[*compiler.Program][]*ast.SourceFile
```

**Key Features:**
- Configurable worker count (defaults to `GOMAXPROCS`)
- Channel-based diagnostic collection
- Context-aware execution with cancellation support

### 3. Rule System (`internal/rule`)
**Responsibilities:**
- Defining rule interfaces and contracts
- Managing rule execution context
- Providing AST visitor pattern implementation
- Handling diagnostic generation and source fixes

**Architecture Pattern:** Visitor Pattern
```go
type RuleListeners map[ast.Kind](func(node *ast.Node))
```

**Rule Lifecycle:**
1. Rule registration with specific AST node types
2. Context creation with TypeScript checker and source file
3. AST traversal triggering registered listeners
4. Diagnostic generation

### 4. TypeScript Integration (`shim/*`)
**Responsibilities:**
- Providing Go bindings to typescript-go internal APIs
- Exposing TypeScript compiler functionality
- Managing AST manipulation and type checking

**Architecture Pattern:** Shim/Proxy Pattern
- Uses Go's `//go:linkname` directive for internal API access
- Virtual file system abstraction for testing and caching
- Direct TypeScript AST usage without conversion

**Key Components:**
- `ast`: TypeScript AST node types and utilities
- `checker`: Type checker bindings
- `compiler`: Program and compilation host
- `scanner`: Source text processing utilities

### 5. Rules Implementation (`internal/rules/*`)
**Current Implementation:** 40+ type-aware rules
- Each rule follows consistent interface pattern
- Type-aware analysis using TypeScript checker
- Compatible with typescript-eslint rule specifications

## Architecture Patterns

### 1. Visitor Pattern (Rule System)
Rules register listeners for specific AST node types, enabling efficient single-pass traversal:
```go
func (r *SomeRule) Run(ctx RuleContext) RuleListeners {
    return RuleListeners{
        ast.FunctionDeclaration: r.checkFunction,
        ast.CallExpression: r.checkCall,
    }
}
```

### 2. Worker Pool Pattern (Parallel Processing)
Multiple goroutines process files concurrently with shared TypeScript checker instances:
- Work distribution via channels
- Shared TypeScript program state
- Concurrent diagnostic collection

### 3. Shim Pattern (TypeScript Integration)
Go bindings to internal TypeScript APIs using linkname directives:
- Zero-copy integration with TypeScript compiler
- Access to internal TypeScript functionality
- Type-safe Go interfaces over JavaScript implementations

## Performance Architecture

### Parallel Processing Model
```
Master Thread                 Worker Threads
     │                             │
     ├── Program Creation          │
     ├── File Discovery            │
     ├── Work Distribution ────────┼── File Processing
     │                             ├── Rule Execution
     ├── Diagnostic Collection ◄───┼── AST Traversal
     └── Output Formatting         └── Type Checking
```

### Memory Management
- **Streaming**: Diagnostics processed as they're generated
- **Buffering**: Output buffered for efficient terminal updates
- **Sharing**: TypeScript program state shared across workers
- **Caching**: Virtual file system enables efficient caching

### CPU Optimization
- **Multi-core**: Utilizes all available CPU cores by default
- **Work Stealing**: Dynamic load balancing across workers
- **Native Speed**: Go compilation to native machine code

## Integration Architecture

### Oxlint Integration (Primary)
**Data Flow:**
1. Oxlint discovers files and applies ignore patterns
2. Oxlint invokes tsgolint with file paths and configuration
3. tsgolint processes files and returns structured diagnostics
4. Oxlint formats and displays results

### Standalone Mode (Secondary)
- Direct CLI usage for development and testing
- Compatible output formatting
- Built-in file discovery and configuration resolution

## Cross-cutting Concerns

### Error Handling
- Graceful degradation for TypeScript compilation errors
- Comprehensive error reporting for configuration issues
- Panic recovery in worker goroutines

### Logging & Observability
- Performance profiling support (`-trace`, `-cpuprof`)
- Structured diagnostic output
- Timing information and statistics

### Configuration Management
- TypeScript configuration inheritance
- Rule-specific configuration (future)
- Environment-specific settings

## Architecture Decisions

### 1. Go Implementation
**Decision:** Implement linter in Go rather than TypeScript/JavaScript
**Rationale:**
- Native performance without JavaScript runtime overhead
- Strong concurrency primitives for parallel processing
- Type safety and memory efficiency
- Better integration with systems tools

### 2. Direct TypeScript AST Usage
**Decision:** Use TypeScript AST directly instead of converting to ESTree
**Rationale:**
- Eliminates conversion overhead
- Preserves all TypeScript-specific information
- Enables more precise type-aware analysis
- Reduces memory footprint

### 3. Worker Pool Architecture
**Decision:** Use fixed worker pool rather than fork-per-file
**Rationale:**
- Amortizes TypeScript program creation cost
- Enables efficient checker sharing
- Reduces memory fragmentation
- Better resource utilization

### 4. Shim-based TypeScript Integration
**Decision:** Use linkname directives to access internal typescript-go APIs
**Rationale:**
- Provides access to full TypeScript compiler functionality
- Maintains type safety through Go interfaces
- Enables zero-copy integration
- **Trade-off:** Depends on internal APIs (not recommended for production use)

### 5. Separation from Oxlint
**Decision:** Implement as separate backend rather than integrated component
**Rationale:**
- Clear separation of concerns
- Independent development and testing
- Potential for multiple frontend integrations
- Easier maintenance and debugging

## Development Architecture

### Build System
- Go modules for dependency management
- Git submodules for typescript-go integration
- Automated shim generation via `tools/gen_shims`
- Patch-based TypeScript modifications

### Testing Strategy
- Rule-specific test fixtures
- Integration tests with real TypeScript projects
- Performance benchmarking
- Compatibility testing with typescript-eslint

### Rule Development
- Consistent rule interface pattern
- Comprehensive rule testing framework
- Documentation and examples for rule authors
- Type-aware testing utilities
