package prefer_for_of

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildPreferForOfMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferForOf",
		Description: "Expected a `for-of` loop instead of a `for` loop with this simple iteration.",
		Help:        "Consider using a for-of loop for this simple iteration.",
	}
}

var PreferForOfRule = rule.Rule{
	Name: "prefer-for-of",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Check if a for loop can be converted to for-of
		checkForStatement := func(node *ast.Node) {
			forStmt := node.AsForStatement()
			
			// Must have initializer, condition, and incrementor
			if forStmt.Initializer == nil || forStmt.Condition == nil || forStmt.Incrementor == nil {
				return
			}
			
			// Check if this is a simple numeric for loop
			// Pattern: for (let i = 0; i < array.length; i++)
			if !isSimpleNumericForLoop(forStmt) {
				return
			}
			
			// Get the loop variable name
			loopVarName := getLoopVariableName(forStmt)
			if loopVarName == "" {
				return
			}
			
			// Get the array being iterated
			arrayExpr := getArrayExpression(forStmt)
			if arrayExpr == nil {
				return
			}
			
			// Check if the loop variable is used properly (accessing the main array at least once)
			// and not used for other problematic purposes
			usageCheck := analyzeLoopVariableUsage(forStmt.Statement, loopVarName, arrayExpr)
			if !usageCheck.isUsedForMainArrayAccess || usageCheck.hasProblematicUse {
				return
			}
			
			// Check if the array type is iterable
			if !isIterableType(ctx, arrayExpr) {
				return
			}
			
			// Report the issue
			ctx.ReportRange(
				utils.GetForStatementHeadLoc(ctx.SourceFile, node),
				buildPreferForOfMessage(),
			)
		}

		return rule.RuleListeners{
			ast.KindForStatement: checkForStatement,
		}
	},
}

// Check if this is a simple numeric for loop: for (let i = 0; i < array.length; i++)
func isSimpleNumericForLoop(forStmt *ast.ForStatement) bool {
	// Check initializer: let i = 0
	if !ast.IsVariableDeclarationList(forStmt.Initializer) {
		return false
	}
	
	varDeclList := forStmt.Initializer.AsVariableDeclarationList()
	if len(varDeclList.Declarations.Nodes) != 1 {
		return false
	}
	
	varDecl := varDeclList.Declarations.Nodes[0].AsVariableDeclaration()
	if !ast.IsIdentifier(varDecl.Name()) || varDecl.Initializer == nil {
		return false
	}
	
	// Check if initializer is 0
	if !ast.IsNumericLiteral(varDecl.Initializer) {
		return false
	}
	
	numLit := varDecl.Initializer.AsNumericLiteral()
	if numLit.Text != "0" {
		return false
	}
	
	// Check condition: i < array.length
	if !ast.IsBinaryExpression(forStmt.Condition) {
		return false
	}
	
	binExpr := forStmt.Condition.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindLessThanToken {
		return false
	}
	
	// Left side should be the loop variable
	if !ast.IsIdentifier(binExpr.Left) {
		return false
	}
	
	leftId := binExpr.Left.AsIdentifier()
	varId := varDecl.Name().AsIdentifier()
	if leftId.Text != varId.Text {
		return false
	}
	
	// Right side should be array.length
	if !isArrayLengthExpression(binExpr.Right) {
		return false
	}
	
	// Check incrementor: i++
	if !isSimpleIncrement(forStmt.Incrementor, varId.Text) {
		return false
	}
	
	return true
}

// Check if expression is array.length
func isArrayLengthExpression(expr *ast.Node) bool {
	if !ast.IsPropertyAccessExpression(expr) {
		return false
	}
	
	propAccess := expr.AsPropertyAccessExpression()
	nameNode := propAccess.Name()
	if !ast.IsIdentifier(nameNode) {
		return false
	}
	
	return nameNode.AsIdentifier().Text == "length"
}

// Check if this is i++ or ++i
func isSimpleIncrement(expr *ast.Node, varName string) bool {
	if expr.Kind == ast.KindPostfixUnaryExpression {
		postfix := expr.AsPostfixUnaryExpression()
		if postfix.Operator != ast.KindPlusPlusToken {
			return false
		}
		if !ast.IsIdentifier(postfix.Operand) {
			return false
		}
		return postfix.Operand.AsIdentifier().Text == varName
	}
	
	if expr.Kind == ast.KindPrefixUnaryExpression {
		prefix := expr.AsPrefixUnaryExpression()
		if prefix.Operator != ast.KindPlusPlusToken {
			return false
		}
		if !ast.IsIdentifier(prefix.Operand) {
			return false
		}
		return prefix.Operand.AsIdentifier().Text == varName
	}
	
	return false
}

// Get the loop variable name
func getLoopVariableName(forStmt *ast.ForStatement) string {
	if !ast.IsVariableDeclarationList(forStmt.Initializer) {
		return ""
	}
	
	varDeclList := forStmt.Initializer.AsVariableDeclarationList()
	if len(varDeclList.Declarations.Nodes) != 1 {
		return ""
	}
	
	varDecl := varDeclList.Declarations.Nodes[0].AsVariableDeclaration()
	if !ast.IsIdentifier(varDecl.Name()) {
		return ""
	}
	
	return varDecl.Name().AsIdentifier().Text
}

// Get the array expression from the condition
func getArrayExpression(forStmt *ast.ForStatement) *ast.Node {
	if !ast.IsBinaryExpression(forStmt.Condition) {
		return nil
	}
	
	binExpr := forStmt.Condition.AsBinaryExpression()
	if !isArrayLengthExpression(binExpr.Right) {
		return nil
	}
	
	propAccess := binExpr.Right.AsPropertyAccessExpression()
	return propAccess.Expression
}

type loopVariableUsageInfo struct {
	isUsedForMainArrayAccess bool
	hasProblematicUse        bool
}

// Analyze how the loop variable is used
func analyzeLoopVariableUsage(stmt *ast.Node, loopVarName string, arrayExpr *ast.Node) loopVariableUsageInfo {
	mainArrayText := getNodeText(arrayExpr)
	result := loopVariableUsageInfo{
		isUsedForMainArrayAccess: false,
		hasProblematicUse:        false,
	}
	
	// Walk through all nodes in the loop body
	visitNode(stmt, func(node *ast.Node) {
		if result.hasProblematicUse {
			return // Short circuit once we find a problematic use
		}
		
		// Look for uses of the loop variable
		if ast.IsIdentifier(node) {
			id := node.AsIdentifier()
			if id.Text == loopVarName {
				// Found a use of the loop variable, check the context
				parent := node.Parent
				
				// Check if this is used as an array index: something[loopVar]
				if ast.IsElementAccessExpression(parent) {
					elemAccess := parent.AsElementAccessExpression()
					if elemAccess.ArgumentExpression == node {
						// This is array[loopVar] - check which array
						arrayBeingIndexed := getNodeText(elemAccess.Expression)
						if arrayBeingIndexed == mainArrayText {
							// Loop variable is used to access the main array - this is good!
							result.isUsedForMainArrayAccess = true
						} else {
							// Loop variable is used to index a different array - this is problematic
							result.hasProblematicUse = true
							return
						}
					} else {
						// Loop variable is used in some other context within element access
						result.hasProblematicUse = true
						return
					}
				} else {
					// Loop variable is used in a non-array-indexing context
					// Check if this is part of the for loop declaration/condition/increment
					if !isPartOfForLoopStructure(node) {
						// Loop variable is used outside the for-loop structure (e.g., console.log(i))
						result.hasProblematicUse = true
						return
					}
				}
			}
		}
	})
	
	return result
}

// Check if a node is part of the for-loop structure (initializer, condition, increment)
func isPartOfForLoopStructure(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if ast.IsForStatement(current) {
			forStmt := current.AsForStatement()
			// Check if this identifier is part of the loop structure itself
			if isNodeWithinNode(node, forStmt.Initializer) ||
				isNodeWithinNode(node, forStmt.Condition) ||
				isNodeWithinNode(node, forStmt.Incrementor) {
				return true
			}
			// If we reached the for statement but the node is not in the structure,
			// it must be in the body
			return false
		}
		current = current.Parent
	}
	return false
}

// Check if nodeToFind is within containerNode
func isNodeWithinNode(nodeToFind *ast.Node, containerNode *ast.Node) bool {
	if containerNode == nil || nodeToFind == nil {
		return false
	}
	
	current := nodeToFind
	for current != nil {
		if current == containerNode {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper to get text representation of a node (improved)
func getNodeText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		exprText := getNodeText(propAccess.Expression)
		nameNode := propAccess.Name()
		if ast.IsIdentifier(nameNode) {
			return exprText + "." + nameNode.AsIdentifier().Text
		}
		return exprText + ".<unknown>"
	case ast.KindElementAccessExpression:
		elemAccess := node.AsElementAccessExpression()
		exprText := getNodeText(elemAccess.Expression)
		argText := getNodeText(elemAccess.ArgumentExpression)
		return exprText + "[" + argText + "]"
	case ast.KindStringLiteral:
		return "\"" + node.AsStringLiteral().Text + "\""
	case ast.KindNumericLiteral:
		return node.AsNumericLiteral().Text
	case ast.KindThisKeyword:
		return "this"
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		return getNodeText(callExpr.Expression) + "(...)"
	default:
		// For unknown node types, return a placeholder
		// In a production implementation, we'd handle more cases
		return "<expr:" + string(rune(node.Kind)) + ">"
	}
}

// Simple node visitor - enhanced to cover more cases
func visitNode(node *ast.Node, visitor func(*ast.Node)) {
	if node == nil {
		return
	}
	
	visitor(node)
	
	// Visit children - handle more node types
	switch node.Kind {
	case ast.KindBlock:
		block := node.AsBlock()
		for _, stmt := range block.Statements.Nodes {
			visitNode(stmt, visitor)
		}
	case ast.KindExpressionStatement:
		exprStmt := node.AsExpressionStatement()
		visitNode(exprStmt.Expression, visitor)
	case ast.KindBinaryExpression:
		binExpr := node.AsBinaryExpression()
		visitNode(binExpr.Left, visitor)
		visitNode(binExpr.Right, visitor)
	case ast.KindElementAccessExpression:
		elemAccess := node.AsElementAccessExpression()
		visitNode(elemAccess.Expression, visitor)
		visitNode(elemAccess.ArgumentExpression, visitor)
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		visitNode(propAccess.Expression, visitor)
		// Note: don't visit Name() as it's the property being accessed
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		visitNode(callExpr.Expression, visitor)
		for _, arg := range callExpr.Arguments.Nodes {
			visitNode(arg, visitor)
		}
	case ast.KindForOfStatement, ast.KindForInStatement:
		forOfStmt := node.AsForInOrOfStatement()
		visitNode(forOfStmt.Initializer, visitor)
		visitNode(forOfStmt.Expression, visitor)
		visitNode(forOfStmt.Statement, visitor)
	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		visitNode(forStmt.Initializer, visitor)
		visitNode(forStmt.Condition, visitor)
		visitNode(forStmt.Incrementor, visitor)
		visitNode(forStmt.Statement, visitor)
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		visitNode(varStmt.DeclarationList, visitor)
	case ast.KindVariableDeclarationList:
		varDeclList := node.AsVariableDeclarationList()
		for _, decl := range varDeclList.Declarations.Nodes {
			visitNode(decl, visitor)
		}
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		visitNode(varDecl.Name(), visitor)
		if varDecl.Initializer != nil {
			visitNode(varDecl.Initializer, visitor)
		}
	case ast.KindPostfixUnaryExpression:
		postfix := node.AsPostfixUnaryExpression()
		visitNode(postfix.Operand, visitor)
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		visitNode(prefix.Operand, visitor)
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		visitNode(paren.Expression, visitor)
	case ast.KindAsExpression:
		asExpr := node.AsAsExpression()
		visitNode(asExpr.Expression, visitor)
	case ast.KindNonNullExpression:
		nonNull := node.AsNonNullExpression()
		visitNode(nonNull.Expression, visitor)
	}
}

// Check if the type is iterable (has array-like characteristics)
func isIterableType(ctx rule.RuleContext, expr *ast.Node) bool {
	t := ctx.TypeChecker.GetTypeAtLocation(expr)
	
	// Check if it's an array-like type
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		// Check for number index signature (arrays have this)
		return utils.GetNumberIndexType(ctx.TypeChecker, t) != nil
	})
}