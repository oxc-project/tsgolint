package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// BenchmarkCreateProgram benchmarks TypeScript program creation
func BenchmarkCreateProgram(b *testing.B) {
	testCode := `
interface User {
  id: number;
  name: string;
  email: string;
}

class UserService {
  private users: User[] = [];
  
  async addUser(user: User): Promise<void> {
    this.users.push(user);
  }
  
  findUserById(id: number): User | undefined {
    return this.users.find(u => u.id === id);
  }
  
  async fetchUsers(): Promise<User[]> {
    const response = await fetch('/api/users');
    return response.json();
  }
}

function processUsers(users: User[]): string[] {
  return users.map(user => user.name.toUpperCase());
}
`

	rootDir := "/tmp/bench"
	filePath := tspath.ResolvePath(rootDir, "test.ts")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fs := NewOverlayVFSForFile(filePath, testCode)
		host := CreateCompilerHost(rootDir, fs)
		
		_, err := CreateInferredProjectProgram(false, fs, rootDir, host, []string{filePath})
		if err != nil {
			b.Fatalf("couldn't create program: %v", err)
		}
	}
}

// BenchmarkCreateProgramMultipleFiles benchmarks program creation with multiple files
func BenchmarkCreateProgramMultipleFiles(b *testing.B) {
	testFiles := map[string]string{
		"/tmp/bench/user.ts": `
export interface User {
  id: number;
  name: string;
  email: string;
}
`,
		"/tmp/bench/service.ts": `
import { User } from './user';

export class UserService {
  private users: User[] = [];
  
  async addUser(user: User): Promise<void> {
    this.users.push(user);
  }
  
  findUserById(id: number): User | undefined {
    return this.users.find(u => u.id === id);
  }
}
`,
		"/tmp/bench/utils.ts": `
import { User } from './user';

export function processUsers(users: User[]): string[] {
  return users.map(user => user.name.toUpperCase());
}

export async function fetchUsers(): Promise<User[]> {
  const response = await fetch('/api/users');
  return response.json();
}
`,
	}

	rootDir := "/tmp/bench"
	var filePaths []string
	for filePath := range testFiles {
		filePaths = append(filePaths, filePath)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fs := NewOverlayVFSForFiles(testFiles)
		host := CreateCompilerHost(rootDir, fs)
		
		_, err := CreateInferredProjectProgram(false, fs, rootDir, host, filePaths)
		if err != nil {
			b.Fatalf("couldn't create program: %v", err)
		}
	}
}

// BenchmarkIsSymbolFromDefaultLibrary benchmarks symbol checking operations
func BenchmarkIsSymbolFromDefaultLibrary(b *testing.B) {
	testCode := `type Test = Array<number>;`
	
	rootDir := "/tmp/bench"
	filePath := tspath.ResolvePath(rootDir, "test.ts")
	fs := NewOverlayVFSForFile(filePath, testCode)
	
	program, err := CreateInferredProjectProgram(false, fs, rootDir, CreateCompilerHost(rootDir, fs), []string{filePath})
	if err != nil {
		b.Fatalf("couldn't create program: %v", err)
	}
	
	sourceFile := program.GetSourceFile(filePath)
	checker, done := program.GetTypeChecker(b.Context())
	defer done()
	
	typeAliasDecl := sourceFile.Statements.Nodes[0].AsTypeAliasDeclaration()
	symbolType := checker.GetTypeAtLocation(typeAliasDecl.Name())
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = IsSymbolFromDefaultLibrary(program, symbolType.Symbol())
	}
}

// BenchmarkASTTraversal benchmarks AST node processing operations
func BenchmarkASTTraversal(b *testing.B) {
	testCode := `
class TestClass {
  private value: number = 0;
  
  constructor(initialValue: number) {
    this.value = initialValue;
  }
  
  async method1(): Promise<number> {
    return this.value * 2;
  }
  
  method2(callback: (x: number) => void): void {
    callback(this.value);
  }
  
  get currentValue(): number {
    return this.value;
  }
  
  set currentValue(newValue: number) {
    this.value = newValue;
  }
}

function processValue(x: number): Promise<string> {
  return Promise.resolve(x.toString());
}

const arrow = (x: number, y: number) => x + y;
`

	rootDir := "/tmp/bench"
	filePath := tspath.ResolvePath(rootDir, "test.ts")
	fs := NewOverlayVFSForFile(filePath, testCode)
	
	program, err := CreateInferredProjectProgram(false, fs, rootDir, CreateCompilerHost(rootDir, fs), []string{filePath})
	if err != nil {
		b.Fatalf("couldn't create program: %v", err)
	}
	
	sourceFile := program.GetSourceFile(filePath)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nodeCount := 0
		
		// Count statements - simple AST processing
		for _, stmt := range sourceFile.Statements.Nodes {
			nodeCount++
			if ast.IsClassDeclaration(stmt) {
				classDecl := stmt.AsClassDeclaration()
				for range classDecl.Members.Nodes {
					nodeCount++
				}
			}
		}
	}
}

// BenchmarkOverlayVFS benchmarks virtual file system operations
func BenchmarkOverlayVFS(b *testing.B) {
	testFiles := make(map[string]string)
	for i := 0; i < 50; i++ {
		filePath := "/tmp/bench/file" + string(rune('0'+(i%10))) + ".ts"
		testFiles[filePath] = `
export function example` + string(rune('0'+(i%10))) + `(): number {
  return ` + string(rune('0'+(i%10))) + `;
}
`
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fs := NewOverlayVFSForFiles(testFiles)
		
		// Simulate file operations
		for filePath := range testFiles {
			exists := fs.FileExists(filePath)
			if exists {
				_, _ = fs.ReadFile(filePath)
			}
		}
	}
}