package no_unnecessary_type_assertion

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const contextFreeAssertionSource = `
interface Base { id: string }
interface Derived extends Base { extra: number }
declare function query<T extends Base = Base>(key: string): T;
const necessary = query("key") as Derived;
const redundant = query<Derived>("key") as Derived;
declare function identity<T extends Base>(value: T): T;
declare const derived: Derived;
const inferred = identity(derived) as Derived;
declare class Box<T extends Base = Base> { value: T }
const constructed = new Box() as Box<Derived>;
declare function tag<T extends Base = Base>(strings: TemplateStringsArray): T;
const tagged = tag` + "`key`" + ` as Derived;
declare function queryAsync<T extends Base = Base>(key: string): Promise<T>;
async function awaited() { return (await queryAsync("key")) as Derived; }
`

func contextFreeAssertionProgram(t *testing.T) (*ast.SourceFile, *linter.RunLinterOnProgramOptions) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "context-free-assertion.ts")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{filePath: contextFreeAssertionSource})
	host := utils.CreateCompilerHost(rootDir, fs)
	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host, false)
	if err != nil {
		t.Fatal(err)
	}
	file := program.GetSourceFile(filePath)
	if file == nil {
		t.Fatal("source file was not added to the program")
	}
	return file, &linter.RunLinterOnProgramOptions{Program: program, Files: []*ast.SourceFile{file}, Workers: 1}
}

func TestContextFreeCallTypeIsIndependentOfCheckOrder(t *testing.T) {
	for _, contextualFirst := range []bool{false, true} {
		name := "contextFreeFirst"
		if contextualFirst {
			name = "contextualFirst"
		}
		t.Run(name, func(t *testing.T) {
			file, options := contextFreeAssertionProgram(t)
			c, done := options.Program.GetTypeChecker(t.Context())
			defer done()
			declaration := file.Statements.Nodes[3].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0]
			call := declaration.Initializer().Expression()

			if contextualFirst {
				if got := c.TypeToString(c.GetTypeAtLocation(call)); got != "Derived" {
					t.Fatalf("contextual type before context-free check = %s, want Derived", got)
				}
			}
			contextFree := c.TypeToString(checker.Checker_getContextFreeTypeOfExpression(c, call))
			if contextFree == "Derived" {
				t.Fatalf("context-free type inherited the assertion: %s", contextFree)
			}
			if got := c.TypeToString(c.GetTypeAtLocation(call)); got != "Derived" {
				t.Fatalf("contextual type after context-free check = %s, want Derived", got)
			}
		})
	}
}

func TestContextFreeCallWithSemanticDiagnostics(t *testing.T) {
	for _, reportSemantic := range []bool{false, true} {
		name := "withoutSemanticDiagnostics"
		if reportSemantic {
			name = "withSemanticDiagnostics"
		}
		t.Run(name, func(t *testing.T) {
			_, options := contextFreeAssertionProgram(t)
			var diagnostics []rule.RuleDiagnostic
			options.GetRulesForFile = func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name: NoUnnecessaryTypeAssertionRule.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoUnnecessaryTypeAssertionRule.Run(ctx, nil)
					},
				}}
			}
			options.OnDiagnostic = func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }
			options.OnInternalDiagnostic = func(d diagnostic.Internal) {}
			options.TypeErrors = linter.TypeErrors{ReportSemantic: reportSemantic}
			if err := linter.RunLinterOnProgram(*options); err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 2 || diagnostics[0].Message.Id != "unnecessaryAssertion" || diagnostics[1].Message.Id != "unnecessaryAssertion" {
				t.Fatalf("got %v rule diagnostics, want two redundant assertions", diagnostics)
			}
			for i, expected := range []string{"const redundant", "const inferred"} {
				lineStart := strings.LastIndex(contextFreeAssertionSource[:diagnostics[i].Range.Pos()], "\n") + 1
				if !strings.HasPrefix(contextFreeAssertionSource[lineStart:], expected) {
					t.Fatalf("diagnostic %d is on %q, want %q", i, contextFreeAssertionSource[lineStart:], expected)
				}
			}
		})
	}
}
