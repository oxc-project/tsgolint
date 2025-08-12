package inefficient_string_concat

import (
	"go/ast"

	"github.com/typescript-eslint/tsgolint/internal/gorule"
)

func buildInefficientStringConcatMessage() gorule.GoRuleMessage {
	return gorule.GoRuleMessage{
		Id:          "inefficientStringConcat",
		Description: "Inefficient string concatenation in loop. Consider using strings.Builder for better performance.",
	}
}

func buildUseStringBuilderMessage() gorule.GoRuleMessage {
	return gorule.GoRuleMessage{
		Id:          "useStringBuilder",
		Description: "Use strings.Builder for string concatenation in loops.",
	}
}

var InefficientStringConcatRule = gorule.GoRule{
	Name: "inefficient-string-concat",
	Run: func(ctx gorule.GoRuleContext, options any) gorule.GoRuleListeners {
		return gorule.GoRuleListeners{
			"ForStmt": func(node ast.Node) {
				checkForStringConcatInLoop(node, ctx)
			},
			"RangeStmt": func(node ast.Node) {
				checkForStringConcatInLoop(node, ctx)
			},
		}
	},
}

func checkForStringConcatInLoop(loopNode ast.Node, ctx gorule.GoRuleContext) {
	var loopBody *ast.BlockStmt
	
	switch stmt := loopNode.(type) {
	case *ast.ForStmt:
		loopBody = stmt.Body
	case *ast.RangeStmt:
		loopBody = stmt.Body
	default:
		return
	}

	if loopBody == nil {
		return
	}

	// Look for string concatenation operations in the loop body
	ast.Inspect(loopBody, func(node ast.Node) bool {
		if assignStmt, ok := node.(*ast.AssignStmt); ok {
			// Check for += with string concatenation
			if len(assignStmt.Lhs) == 1 && len(assignStmt.Rhs) == 1 {
				if assignStmt.Tok.String() == "+=" {
					// Check if this is likely string concatenation
					if isLikelyStringVar(assignStmt.Lhs[0]) || isLikelyStringVar(assignStmt.Rhs[0]) {
						ctx.ReportNodeWithSuggestions(
							assignStmt,
							buildInefficientStringConcatMessage(),
							gorule.GoRuleSuggestion{
								Message:  buildUseStringBuilderMessage(),
								FixesArr: []gorule.GoRuleFix{}, // We could add auto-fix later
							},
						)
					}
				}
			}
		}

		// Check for direct assignment with binary + operator
		if assignStmt, ok := node.(*ast.AssignStmt); ok {
			for _, rhs := range assignStmt.Rhs {
				if binExpr, ok := rhs.(*ast.BinaryExpr); ok {
					if binExpr.Op.String() == "+" {
						// Check if either operand is likely a string
						if isLikelyStringVar(binExpr.X) || isLikelyStringVar(binExpr.Y) {
							// Check if the left-hand side variable is being reassigned to itself
							for _, lhs := range assignStmt.Lhs {
								if ident, ok := lhs.(*ast.Ident); ok {
									if identX, ok := binExpr.X.(*ast.Ident); ok && identX.Name == ident.Name {
										ctx.ReportNodeWithSuggestions(
											assignStmt,
											buildInefficientStringConcatMessage(),
											gorule.GoRuleSuggestion{
												Message:  buildUseStringBuilderMessage(),
												FixesArr: []gorule.GoRuleFix{},
											},
										)
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})
}

// isLikelyStringVar checks if a node is likely a string variable
func isLikelyStringVar(node ast.Expr) bool {
	switch n := node.(type) {
	case *ast.Ident:
		// Simple heuristic: variable names containing "str", "msg", "text", etc.
		name := n.Name
		return name == "s" || name == "str" || name == "msg" || name == "text" || 
			   name == "result" || name == "output" || name == "buffer"
	case *ast.BasicLit:
		// String literals
		return n.Kind.String() == "STRING"
	default:
		return false
	}
}