package no_unsafe_enum_comparison

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/jsnum"
)

type staticValueKind uint8

const (
	staticValueString staticValueKind = iota
	staticValueNumber
)

type staticValue struct {
	kind        staticValueKind
	stringValue string
	numberValue jsnum.Number
}

func getStaticValue(node *ast.Node) (staticValue, bool) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return staticValue{}, false
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return staticValue{kind: staticValueString, stringValue: node.Text()}, true
	case ast.KindNumericLiteral:
		return staticValue{kind: staticValueNumber, numberValue: jsnum.FromString(node.Text())}, true
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		value, ok := getStaticValue(prefix.Operand)
		if !ok || value.kind != staticValueNumber {
			return staticValue{}, false
		}
		switch prefix.Operator {
		case ast.KindMinusToken:
			return staticValue{kind: staticValueNumber, numberValue: -value.numberValue}, true
		case ast.KindPlusToken:
			return value, true
		}
	case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
		return getStaticValue(node.Expression())
	case ast.KindBinaryExpression:
		expr := node.AsBinaryExpression()
		if expr.OperatorToken.Kind != ast.KindPlusToken {
			return staticValue{}, false
		}

		left, leftOk := getStaticValue(expr.Left)
		right, rightOk := getStaticValue(expr.Right)
		if !leftOk || !rightOk {
			return staticValue{}, false
		}
		if left.kind == staticValueString || right.kind == staticValueString {
			return staticValue{kind: staticValueString, stringValue: staticValueToString(left) + staticValueToString(right)}, true
		}
		return staticValue{kind: staticValueNumber, numberValue: left.numberValue + right.numberValue}, true
	}

	return staticValue{}, false
}

func staticValueToString(value staticValue) string {
	if value.kind == staticValueString {
		return value.stringValue
	}
	return value.numberValue.String()
}
