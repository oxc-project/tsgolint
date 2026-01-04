package only_throw_error

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "object",
		Description: "Expected an error object to be thrown.",
	}
}
func buildUndefMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "undef",
		Description: "Do not throw undefined.",
	}
}

// isRethrownError checks if the node is a rethrown caught error.
// This handles:
// 1. try { } catch (e) { throw e; }
// 2. promise.catch(e => { throw e; })
// 3. promise.then(onFulfilled, e => { throw e; })
func isRethrownError(ctx rule.RuleContext, node *ast.Node) bool {
	if !ast.IsIdentifier(node) {
		return false
	}

	// Get the declaration of the variable
	decl := utils.GetDeclaration(ctx.TypeChecker, node)
	if decl == nil {
		return false
	}

	// Case 1: try { } catch (e) { throw e; }
	if ast.IsCatchClause(decl.Parent) {
		return true
	}

	// Case 2 & 3: promise.catch(e => { throw e; }) or promise.then(onFulfilled, e => { throw e; })
	// The declaration must be from a parameter of an arrow function
	if !ast.IsParameter(decl) {
		return false
	}

	paramDecl := decl.AsParameterDeclaration()

	// Must not be a rest parameter (...e)
	if paramDecl.DotDotDotToken != nil {
		return false
	}

	// The parameter must belong to an arrow function
	funcNode := decl.Parent
	if !ast.IsArrowFunction(funcNode) {
		return false
	}

	arrowFunc := funcNode.AsArrowFunction()

	// The parameter must be the first parameter
	if len(arrowFunc.Parameters.Nodes) == 0 || arrowFunc.Parameters.Nodes[0] != decl {
		return false
	}

	// The arrow function must be a direct argument of a call expression
	if funcNode.Parent == nil || !ast.IsCallExpression(funcNode.Parent) {
		return false
	}

	callExpr := funcNode.Parent.AsCallExpression()

	// Check if this is a .catch() or .then() call
	if !ast.IsPropertyAccessExpression(callExpr.Expression) {
		return false
	}

	propAccess := callExpr.Expression.AsPropertyAccessExpression()
	methodName := propAccess.Name().Text()

	// For .catch(e => { throw e; }), the arrow function must be the first argument (onRejected)
	// For .then(onFulfilled, e => { throw e; }), the arrow function must be the second argument (onRejected)
	isRejectionHandler := false
	args := callExpr.Arguments.Nodes

	if methodName == "catch" {
		// .catch(onRejected)
		// First argument must be our arrow function and not preceded by spread
		if len(args) >= 1 && args[0] == funcNode && !ast.IsSpreadElement(args[0]) {
			isRejectionHandler = true
		}
	} else if methodName == "then" {
		// .then(onFulfilled, onRejected)
		// Second argument must be our arrow function
		// Also check that neither first nor second argument is a spread element
		if len(args) >= 2 && args[1] == funcNode {
			if !ast.IsSpreadElement(args[0]) && !ast.IsSpreadElement(args[1]) {
				isRejectionHandler = true
			}
		}
	}

	if !isRejectionHandler {
		return false
	}

	// Verify that the object is actually a thenable (Promise)
	objectNode := propAccess.Expression
	objectType := ctx.TypeChecker.GetTypeAtLocation(objectNode)
	if !utils.IsThenableType(ctx.TypeChecker, objectNode, objectType) {
		return false
	}

	return true
}

var OnlyThrowErrorRule = rule.Rule{
	Name: "only-throw-error",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[OnlyThrowErrorOptions](options, "only-throw-error")

		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				expr := node.Expression()
				t := ctx.TypeChecker.GetTypeAtLocation(expr)

				if utils.TypeMatchesSomeSpecifier(t, opts.Allow, ctx.Program) {
					return
				}

				if utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
					ctx.ReportNode(node, buildUndefMessage())
					return
				}

				if opts.AllowThrowingAny && utils.IsTypeAnyType(t) {
					return
				}

				if opts.AllowThrowingUnknown && utils.IsTypeUnknownType(t) {
					return
				}

				if opts.AllowRethrowing && isRethrownError(ctx, expr) {
					return
				}

				if utils.IsErrorLike(ctx.Program, ctx.TypeChecker, t) {
					return
				}

				ctx.ReportNode(expr, buildObjectMessage())
			},
		}
	},
}
