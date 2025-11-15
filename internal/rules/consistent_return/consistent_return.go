package consistent_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMissingReturnValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnValue",
		Description: "Expected to return a value at the end of this function.",
	}
}

func buildUnexpectedReturnValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedReturnValue",
		Description: "Function should not return a value.",
	}
}

type functionInfo struct {
	hasReturnWithValue           bool
	hasReturnWithoutValue        bool
	returnStatementsWithValue    []*ast.Node
	returnStatementsWithoutValue []*ast.Node
}

var ConsistentReturnRule = rule.Rule{
	Name: "consistent-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[ConsistentReturnOptions](options, "consistent-return")

		functionStack := []*functionInfo{}

		isReturningVoid := func(node *ast.Node) bool {
			if !ast.IsFunctionLikeDeclaration(node) {
				return false
			}

			functionType := ctx.TypeChecker.GetTypeAtLocation(node)
			callSignatures := utils.CollectAllCallSignatures(ctx.TypeChecker, functionType)

			for _, signature := range callSignatures {
				returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)

				// For async functions, unwrap the Promise type
				if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
					awaitedType := checker.Checker_getAwaitedType(ctx.TypeChecker, returnType)
					if awaitedType != nil {
						returnType = awaitedType
					}
				}

				// Check if the return type is void
				// Note: We only check for void, not undefined
				// undefined is a specific value type, not "no value"
				if utils.IsTypeFlagSet(returnType, checker.TypeFlagsVoid) {
					return true
				}

				// Also check if it's a union type containing void
				if returnType.Flags()&checker.TypeFlagsUnion != 0 {
					for _, typePart := range utils.UnionTypeParts(returnType) {
						if utils.IsTypeFlagSet(typePart, checker.TypeFlagsVoid) {
							return true
						}
					}
				}
			}

			return false
		}

		isReturningUndefined := func(returnNode *ast.Node) bool {
			if returnNode == nil {
				return false
			}

			returnType := ctx.TypeChecker.GetTypeAtLocation(returnNode)

			// Check if it's explicitly undefined type
			return utils.IsTypeFlagSet(returnType, checker.TypeFlagsUndefined)
		}

		hasReturnValue := func(returnStmt *ast.Node) bool {
			expr := returnStmt.Expression()
			if expr == nil {
				return false
			}

			// With treatUndefinedAsUnspecified option, treat undefined as no value
			if opts.TreatUndefinedAsUnspecified && isReturningUndefined(expr) {
				return false
			}

			return true
		}

		enterFunction := func(node *ast.Node) {
			functionStack = append(functionStack, &functionInfo{
				returnStatementsWithValue:    []*ast.Node{},
				returnStatementsWithoutValue: []*ast.Node{},
			})
		}

		exitFunction := func(node *ast.Node) {
			if len(functionStack) == 0 {
				return
			}

			info := functionStack[len(functionStack)-1]
			functionStack = functionStack[:len(functionStack)-1]

			// If function explicitly returns void, we don't need to check consistency
			if isReturningVoid(node) {
				return
			}

			// Check for inconsistent returns
			if info.hasReturnWithValue && info.hasReturnWithoutValue {
				// Determine which type of error to report based on the first return statement
				// If the first return had no value, report unexpected value on those with values
				// Otherwise, report missing value on those without values

				firstReturnWithoutValue := (*ast.Node)(nil)
				if len(info.returnStatementsWithoutValue) > 0 {
					firstReturnWithoutValue = info.returnStatementsWithoutValue[0]
				}

				firstReturnWithValue := (*ast.Node)(nil)
				if len(info.returnStatementsWithValue) > 0 {
					firstReturnWithValue = info.returnStatementsWithValue[0]
				}

				// Determine which came first
				if firstReturnWithoutValue != nil && firstReturnWithValue != nil {
					if firstReturnWithoutValue.Pos() < firstReturnWithValue.Pos() {
						// First return had no value, so returns with values are unexpected
						for _, stmt := range info.returnStatementsWithValue {
							ctx.ReportNode(stmt, buildUnexpectedReturnValueMessage())
						}
					} else {
						// First return had a value, so returns without values are missing values
						for _, stmt := range info.returnStatementsWithoutValue {
							ctx.ReportNode(stmt, buildMissingReturnValueMessage())
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindArrowFunction:                            enterFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):       exitFunction,
			ast.KindMethodDeclaration:                        enterFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,

			ast.KindReturnStatement: func(node *ast.Node) {
				if len(functionStack) == 0 {
					return
				}

				info := functionStack[len(functionStack)-1]

				if hasReturnValue(node) {
					info.hasReturnWithValue = true
					info.returnStatementsWithValue = append(info.returnStatementsWithValue, node)
				} else {
					info.hasReturnWithoutValue = true
					info.returnStatementsWithoutValue = append(info.returnStatementsWithoutValue, node)
				}
			},
		}
	},
}
