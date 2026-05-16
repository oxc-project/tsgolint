package no_object_comparison

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildObjectComparisonMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "objectComparison",
		Description: "Do not compare objects using this operator.",
	}
}

var defaultForbiddenOperators = []string{"<", "<=", ">", ">="}
var equalityOperators = []string{"==", "===", "!=", "!=="}

func operatorText(kind ast.Kind) string {
	switch kind {
	case ast.KindLessThanToken:
		return "<"
	case ast.KindLessThanEqualsToken:
		return "<="
	case ast.KindGreaterThanToken:
		return ">"
	case ast.KindGreaterThanEqualsToken:
		return ">="
	case ast.KindEqualsEqualsToken:
		return "=="
	case ast.KindEqualsEqualsEqualsToken:
		return "==="
	case ast.KindExclamationEqualsToken:
		return "!="
	case ast.KindExclamationEqualsEqualsToken:
		return "!=="
	default:
		return ""
	}
}

func buildForbiddenOperatorsByClass(classes []NoObjectComparisonClassOption) map[string]*utils.Set[string] {
	if len(classes) == 0 {
		return nil
	}

	forbiddenOperatorsByClass := make(map[string]*utils.Set[string], len(classes))
	for _, class := range classes {
		if class.Name == "" {
			continue
		}

		operators := utils.NewSetFromItems(defaultForbiddenOperators...)
		if class.ForbidEqualityOperators {
			for _, operator := range equalityOperators {
				operators.Add(operator)
			}
		}

		forbiddenOperatorsByClass[class.Name] = operators
	}

	return forbiddenOperatorsByClass
}

func forbidsOperator(t *checker.Type, operator string, forbiddenOperatorsByClass map[string]*utils.Set[string]) bool {
	if t == nil || len(forbiddenOperatorsByClass) == 0 {
		return false
	}

	if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
		return utils.Some(t.Types(), func(part *checker.Type) bool {
			return forbidsOperator(part, operator, forbiddenOperatorsByClass)
		})
	}

	if utils.IsTypeFlagSet(t, checker.TypeFlagsObject) && checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0 {
		t = t.Target()
	}

	if !utils.IsObjectType(t) || checker.Type_objectFlags(t)&checker.ObjectFlagsClassOrInterface == 0 {
		return false
	}

	symbol := checker.Type_symbol(t)
	if symbol == nil {
		return false
	}

	forbiddenOperators := forbiddenOperatorsByClass[symbol.Name]
	return forbiddenOperators != nil && forbiddenOperators.Has(operator)
}

var NoObjectComparisonRule = rule.Rule{
	Name: "no-object-comparison",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoObjectComparisonOptions](options, "no-object-comparison")
		forbiddenOperatorsByClass := buildForbiddenOperatorsByClass(opts.Classes)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				operator := operatorText(expr.OperatorToken.Kind)
				if operator == "" {
					return
				}

				if utils.IsNullLiteralOrUndefinedIdentifier(expr.Left) || utils.IsNullLiteralOrUndefinedIdentifier(expr.Right) {
					return
				}

				leftType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Left)
				rightType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Right)

				if forbidsOperator(leftType, operator, forbiddenOperatorsByClass) ||
					forbidsOperator(rightType, operator, forbiddenOperatorsByClass) {
					ctx.ReportNode(node, buildObjectComparisonMessage())
				}
			},
		}
	},
}
