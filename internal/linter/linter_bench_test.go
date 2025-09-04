package linter

import (
	"context"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_floating_promises"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// BenchmarkLinterOnSingleFile benchmarks linting a single TypeScript file
func BenchmarkLinterOnSingleFile(b *testing.B) {
	// Create a test TypeScript file content
	testCode := `
async function example() {
  const promise = fetch('/api/data');
  const result = await promise;
  return result.json();
}

function floatingPromise() {
  fetch('/api/data'); // This should trigger no-floating-promises
  Promise.resolve(42);
}

class TestClass {
  async method() {
    await this.asyncOperation();
  }
  
  asyncOperation(): Promise<void> {
    return Promise.resolve();
  }
}
`

	rootDir := "/tmp/bench"
	filePath := tspath.ResolvePath(rootDir, "test.ts")
	fs := utils.NewOverlayVFSForFile(filePath, testCode)
	
	// Create program once
	program, err := utils.CreateInferredProjectProgram(false, fs, rootDir, utils.CreateCompilerHost(rootDir, fs), []string{filePath})
	if err != nil {
		b.Fatalf("couldn't create program: %v", err)
	}
	
	sourceFile := program.GetSourceFile(filePath)
	
	rules := []ConfiguredRule{
		{
			Name: "no-floating-promises",
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return no_floating_promises.NoFloatingPromisesRule.Run(ctx, no_floating_promises.NoFloatingPromisesOptions{})
			},
		},
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		diagnosticCount := 0
		onDiagnostic := func(diagnostic rule.RuleDiagnostic) {
			diagnosticCount++
		}
		
		err := RunLinterOnProgram(program, []*ast.SourceFile{sourceFile}, 1, func(*ast.SourceFile) []ConfiguredRule {
			return rules
		}, onDiagnostic)
		
		if err != nil {
			b.Fatalf("linting failed: %v", err)
		}
	}
}

// BenchmarkLinterParallel benchmarks parallel linting
func BenchmarkLinterParallel(b *testing.B) {
	// Create multiple test files
	testFiles := make(map[string]string)
	for i := 0; i < 10; i++ {
		testFiles[tspath.ResolvePath("/tmp/bench", "test"+string(rune('0'+i))+".ts")] = `
async function example` + string(rune('0'+i)) + `() {
  const promise = fetch('/api/data');
  Promise.resolve(42); // floating promise
  const result = await promise;
  return result.json();
}
`
	}

	rootDir := "/tmp/bench"
	fs := utils.NewOverlayVFSForFiles(testFiles)
	
	var filePaths []string
	for filePath := range testFiles {
		filePaths = append(filePaths, filePath)
	}
	
	program, err := utils.CreateInferredProjectProgram(false, fs, rootDir, utils.CreateCompilerHost(rootDir, fs), filePaths)
	if err != nil {
		b.Fatalf("couldn't create program: %v", err)
	}
	
	sourceFiles := program.SourceFiles()
	
	rules := []ConfiguredRule{
		{
			Name: "no-floating-promises",
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return no_floating_promises.NoFloatingPromisesRule.Run(ctx, no_floating_promises.NoFloatingPromisesOptions{})
			},
		},
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		diagnosticCount := 0
		onDiagnostic := func(diagnostic rule.RuleDiagnostic) {
			diagnosticCount++
		}
		
		err := RunLinterOnProgram(program, sourceFiles, 4, func(*ast.SourceFile) []ConfiguredRule {
			return rules
		}, onDiagnostic)
		
		if err != nil {
			b.Fatalf("linting failed: %v", err)
		}
	}
}

// BenchmarkTypeChecking benchmarks TypeScript type checking operations
func BenchmarkTypeChecking(b *testing.B) {
	testCode := `
interface User {
  id: number;
  name: string;
  email: string;
}

function processUser(user: User): Promise<string> {
  return Promise.resolve(user.name.toUpperCase());
}

async function fetchUsers(): Promise<User[]> {
  const response = await fetch('/api/users');
  return response.json();
}

class UserService {
  private users: User[] = [];
  
  async addUser(user: User): Promise<void> {
    this.users.push(user);
  }
  
  findUserById(id: number): User | undefined {
    return this.users.find(u => u.id === id);
  }
}
`

	rootDir := "/tmp/bench"
	filePath := tspath.ResolvePath(rootDir, "test.ts")
	fs := utils.NewOverlayVFSForFile(filePath, testCode)
	
	program, err := utils.CreateInferredProjectProgram(false, fs, rootDir, utils.CreateCompilerHost(rootDir, fs), []string{filePath})
	if err != nil {
		b.Fatalf("couldn't create program: %v", err)
	}
	
	sourceFile := program.GetSourceFile(filePath)

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		checker, done := program.GetTypeChecker(core.WithRequestID(context.Background(), "bench"))
		
		// Perform type checking on various nodes
		for _, stmt := range sourceFile.Statements.Nodes {
			if ast.IsFunctionDeclaration(stmt) {
				funcDecl := stmt.AsFunctionDeclaration()
				if funcDecl.Name() != nil {
					_ = checker.GetTypeAtLocation(funcDecl.Name())
				}
			}
		}
		
		done()
	}
}