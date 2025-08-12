package golinter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/typescript-eslint/tsgolint/internal/gorule"
)

type ConfiguredGoRule struct {
	Name string
	Run  func(ctx gorule.GoRuleContext) gorule.GoRuleListeners
}

type GoFile struct {
	File    *ast.File
	FileSet *token.FileSet
	Path    string
}

// ParseGoFiles parses Go files from the given directory
func ParseGoFiles(rootDir string) ([]GoFile, error) {
	var goFiles []GoFile
	fileSet := token.NewFileSet()

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, node_modules, and hidden directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "node_modules" || name == ".git" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a Go file
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files for now (they often have different patterns)
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files with parse errors
			return nil
		}

		goFiles = append(goFiles, GoFile{
			File:    file,
			FileSet: fileSet,
			Path:    path,
		})

		return nil
	})

	return goFiles, err
}

// RunGoLinter runs Go linting rules on the provided files
func RunGoLinter(
	files []GoFile,
	getRulesForFile func(file GoFile) []ConfiguredGoRule,
	onDiagnostic func(diagnostic gorule.GoRuleDiagnostic),
) error {
	for _, file := range files {
		rules := getRulesForFile(file)

		// Process each rule separately to track rule names
		for _, rule := range rules {
			// Create rule context with rule name
			ctx := gorule.GoRuleContext{
				File:    file.File,
				FileSet: file.FileSet,
				ReportRange: func(start, end token.Pos, msg gorule.GoRuleMessage) {
					onDiagnostic(gorule.GoRuleDiagnostic{
						Range:    start,
						End:      end,
						RuleName: rule.Name,
						Message:  msg,
						FileName: file.Path,
					})
				},
				ReportRangeWithSuggestions: func(start, end token.Pos, msg gorule.GoRuleMessage, suggestions ...gorule.GoRuleSuggestion) {
					onDiagnostic(gorule.GoRuleDiagnostic{
						Range:       start,
						End:         end,
						RuleName:    rule.Name,
						Message:     msg,
						Suggestions: &suggestions,
						FileName:    file.Path,
					})
				},
				ReportNode: func(node ast.Node, msg gorule.GoRuleMessage) {
					onDiagnostic(gorule.GoRuleDiagnostic{
						Range:    node.Pos(),
						End:      node.End(),
						RuleName: rule.Name,
						Message:  msg,
						FileName: file.Path,
					})
				},
				ReportNodeWithFixes: func(node ast.Node, msg gorule.GoRuleMessage, fixes ...gorule.GoRuleFix) {
					onDiagnostic(gorule.GoRuleDiagnostic{
						Range:    node.Pos(),
						End:      node.End(),
						RuleName: rule.Name,
						Message:  msg,
						FixesPtr: &fixes,
						FileName: file.Path,
					})
				},
				ReportNodeWithSuggestions: func(node ast.Node, msg gorule.GoRuleMessage, suggestions ...gorule.GoRuleSuggestion) {
					onDiagnostic(gorule.GoRuleDiagnostic{
						Range:       node.Pos(),
						End:         node.End(),
						RuleName:    rule.Name,
						Message:     msg,
						Suggestions: &suggestions,
						FileName:    file.Path,
					})
				},
			}

			// Get listeners for this rule
			listeners := rule.Run(ctx)

			// Visit AST and call listeners for this rule
			ast.Inspect(file.File, func(node ast.Node) bool {
				if node == nil {
					return false
				}

				nodeType := getNodeTypeName(node)
				if listener, exists := listeners[nodeType]; exists {
					listener(node)
				}

				return true
			})
		}
	}

	return nil
}

// getNodeTypeName returns the type name of an AST node
func getNodeTypeName(node ast.Node) string {
	switch node.(type) {
	case *ast.FuncDecl:
		return "FuncDecl"
	case *ast.GenDecl:
		return "GenDecl"
	case *ast.ValueSpec:
		return "ValueSpec"
	case *ast.TypeSpec:
		return "TypeSpec"
	case *ast.ImportSpec:
		return "ImportSpec"
	case *ast.AssignStmt:
		return "AssignStmt"
	case *ast.IfStmt:
		return "IfStmt"
	case *ast.ForStmt:
		return "ForStmt"
	case *ast.RangeStmt:
		return "RangeStmt"
	case *ast.SwitchStmt:
		return "SwitchStmt"
	case *ast.TypeSwitchStmt:
		return "TypeSwitchStmt"
	case *ast.SelectStmt:
		return "SelectStmt"
	case *ast.CallExpr:
		return "CallExpr"
	case *ast.FuncLit:
		return "FuncLit"
	case *ast.CompositeLit:
		return "CompositeLit"
	case *ast.Ident:
		return "Ident"
	case *ast.BasicLit:
		return "BasicLit"
	case *ast.BinaryExpr:
		return "BinaryExpr"
	case *ast.UnaryExpr:
		return "UnaryExpr"
	case *ast.BlockStmt:
		return "BlockStmt"
	case *ast.ReturnStmt:
		return "ReturnStmt"
	case *ast.DeferStmt:
		return "DeferStmt"
	case *ast.GoStmt:
		return "GoStmt"
	default:
		return "Unknown"
	}
}