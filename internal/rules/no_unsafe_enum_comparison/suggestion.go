package no_unsafe_enum_comparison

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/jsnum"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildReplaceValueWithEnumMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceValueWithEnum",
		Description: "Replace with an enum value comparison.",
	}
}

func enumValueMatchesStaticValue(enumValue any, value staticValue) bool {
	switch value.kind {
	case staticValueString:
		enumString, ok := enumValue.(string)
		return ok && enumString == value.stringValue
	case staticValueNumber:
		enumNumber, ok := enumValue.(jsnum.Number)
		return ok && enumNumber == value.numberValue
	default:
		return false
	}
}

func enumMemberSuffixText(sourceFile *ast.SourceFile, member *ast.Node) string {
	memberName := member.Name()
	if memberName == nil {
		return ""
	}

	switch memberName.Kind {
	case ast.KindIdentifier:
		return "." + memberName.AsIdentifier().Text
	case ast.KindStringLiteral:
		text := memberName.AsStringLiteral().Text
		if scanner.IsIdentifierText(text, core.LanguageVariantStandard) {
			return "." + text
		}
		return "[" + utils.QuoteSingleStringLiteral(text) + "]"
	case ast.KindNoSubstitutionTemplateLiteral:
		text := memberName.AsNoSubstitutionTemplateLiteral().Text
		if scanner.IsIdentifierText(text, core.LanguageVariantStandard) {
			return "." + text
		}
		return "[" + utils.QuoteSingleStringLiteral(text) + "]"
	case ast.KindComputedPropertyName:
		expressionRange := utils.TrimNodeTextRange(sourceFile, memberName.AsComputedPropertyName().Expression)
		return "[" + sourceFile.Text()[expressionRange.Pos():expressionRange.End()] + "]"
	default:
		return ""
	}
}

func isQualifiedIdentifierText(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if !scanner.IsIdentifierText(part, core.LanguageVariantStandard) {
			return false
		}
	}
	return true
}

func symbolMatchesEnum(typeChecker *checker.Checker, symbol *ast.Symbol, enumSymbol *ast.Symbol) bool {
	if symbol == nil || enumSymbol == nil {
		return false
	}
	if symbol == enumSymbol || typeChecker.GetExportSymbolOfSymbol(symbol) == enumSymbol {
		return true
	}
	if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
		aliased := typeChecker.GetAliasedSymbol(symbol)
		return aliased == enumSymbol || typeChecker.GetExportSymbolOfSymbol(aliased) == enumSymbol
	}
	return false
}

func enumQualifierFromType(typeChecker *checker.Checker, atNode *ast.Node, enumType *checker.Type, enumDeclaration *ast.Node) string {
	qualifier := typeChecker.TypeToString(enumType)
	if !isQualifiedIdentifierText(qualifier) {
		return ""
	}

	if strings.Contains(qualifier, ".") {
		return qualifier
	}

	enumSymbol := typeChecker.GetSymbolAtLocation(enumDeclaration.AsEnumDeclaration().Name())
	for _, scopeSymbol := range typeChecker.GetSymbolsInScope(atNode, ast.SymbolFlagsValue) {
		if scopeSymbol.Name == qualifier && symbolMatchesEnum(typeChecker, scopeSymbol, enumSymbol) {
			return qualifier
		}
	}

	return ""
}

func enumComparisonSuggestions(
	sourceFile *ast.SourceFile,
	typeChecker *checker.Checker,
	node *ast.Node,
	expr *ast.BinaryExpression,
	leftType *checker.Type,
	rightType *checker.Type,
) []rule.RuleSuggestion {
	if rightValue, ok := getStaticValue(expr.Right); ok {
		if enumKey := getEnumKeyForLiteral(sourceFile, typeChecker, node, expr.Left, utils.GetEnumLiterals(leftType), rightValue); enumKey != "" {
			return []rule.RuleSuggestion{{
				Message:  buildReplaceValueWithEnumMessage(),
				FixesArr: []rule.RuleFix{rule.RuleFixReplace(sourceFile, expr.Right, enumKey)},
			}}
		}
	}

	if leftValue, ok := getStaticValue(expr.Left); ok {
		if enumKey := getEnumKeyForLiteral(sourceFile, typeChecker, node, expr.Right, utils.GetEnumLiterals(rightType), leftValue); enumKey != "" {
			return []rule.RuleSuggestion{{
				Message:  buildReplaceValueWithEnumMessage(),
				FixesArr: []rule.RuleFix{rule.RuleFixReplace(sourceFile, expr.Left, enumKey)},
			}}
		}
	}

	return []rule.RuleSuggestion{}
}

func symbolMatchesEnumDeclaration(typeChecker *checker.Checker, symbol *ast.Symbol, enumDeclaration *ast.Node) bool {
	if symbol == nil || enumDeclaration == nil {
		return false
	}

	if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
		symbol = typeChecker.GetAliasedSymbol(symbol)
	}
	symbol = typeChecker.GetExportSymbolOfSymbol(symbol)
	return symbol != nil && symbol.ValueDeclaration != nil && ast.IsEnumMember(symbol.ValueDeclaration) && symbol.ValueDeclaration.Parent == enumDeclaration
}

func enumQualifierFromMemberAccess(sourceFile *ast.SourceFile, typeChecker *checker.Checker, node *ast.Node, enumDeclaration *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node == nil {
		return ""
	}

	if ast.IsPropertyAccessExpression(node) {
		symbol := typeChecker.GetSymbolAtLocation(node.AsPropertyAccessExpression().Name())
		if !symbolMatchesEnumDeclaration(typeChecker, symbol, enumDeclaration) {
			return ""
		}
		qualifierRange := utils.TrimNodeTextRange(sourceFile, node.AsPropertyAccessExpression().Expression)
		return sourceFile.Text()[qualifierRange.Pos():qualifierRange.End()]
	}
	if ast.IsElementAccessExpression(node) {
		symbol := typeChecker.GetSymbolAtLocation(node)
		if !symbolMatchesEnumDeclaration(typeChecker, symbol, enumDeclaration) {
			symbol = typeChecker.GetSymbolAtLocation(node.AsElementAccessExpression().ArgumentExpression)
			if !symbolMatchesEnumDeclaration(typeChecker, symbol, enumDeclaration) {
				return ""
			}
		}
		qualifierRange := utils.TrimNodeTextRange(sourceFile, node.AsElementAccessExpression().Expression)
		return sourceFile.Text()[qualifierRange.Pos():qualifierRange.End()]
	}

	return ""
}

func getEnumKeyForLiteral(sourceFile *ast.SourceFile, typeChecker *checker.Checker, atNode *ast.Node, enumSideNode *ast.Node, enumLiterals []*checker.Type, value staticValue) string {
	for _, enumLiteral := range enumLiterals {
		symbol := checker.Type_symbol(enumLiteral)
		if symbol == nil || symbol.ValueDeclaration == nil || !ast.IsEnumMember(symbol.ValueDeclaration) {
			continue
		}

		enumValue := typeChecker.GetConstantValue(symbol.ValueDeclaration)
		if !enumValueMatchesStaticValue(enumValue, value) {
			continue
		}

		suffix := enumMemberSuffixText(sourceFile, symbol.ValueDeclaration)
		if suffix == "" {
			continue
		}

		enumDeclaration := symbol.ValueDeclaration.Parent
		qualifier := enumQualifierFromMemberAccess(sourceFile, typeChecker, enumSideNode, enumDeclaration)
		if qualifier == "" {
			qualifier = enumQualifierFromType(typeChecker, atNode, typeChecker.GetTypeAtLocation(enumDeclaration), enumDeclaration)
		}
		if qualifier == "" {
			continue
		}

		return qualifier + suffix
	}

	return ""
}
