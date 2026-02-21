package consistent_type_exports

import (
	"slices"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildTypeOverValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "typeOverValue",
		Description: "All exports in the declaration are only used as types. Use `export type`.",
	}
}

func buildSingleExportIsTypeMessage(exportName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "singleExportIsType",
		Description: "Type export " + exportName + " is not a value and should be exported using `export type`.",
	}
}

func buildMultipleExportsAreTypesMessage(exportNames string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "multipleExportsAreTypes",
		Description: "Type exports " + exportNames + " are not values and should be exported using `export type`.",
	}
}

type analyzedNamedExport struct {
	node           *ast.Node
	typeTexts      []string
	typeBasedNodes []*ast.Node
	typeBasedNames []string
	valueTexts     []string
	moduleSource   string
}

func isSymbolTypeBased(typeChecker *checker.Checker, symbol *ast.Symbol) (bool, bool) {
	visited := map[*ast.Symbol]bool{}
	for symbol != nil && !visited[symbol] {
		visited[symbol] = true

		if slices.ContainsFunc(symbol.Declarations, ast.IsTypeOnlyImportOrExportDeclaration) {
			return true, true
		}

		if symbol.Flags&ast.SymbolFlagsValue != 0 {
			return false, true
		}

		if symbol.Flags&ast.SymbolFlagsAlias != 0 {
			next := checker.Checker_getImmediateAliasedSymbol(typeChecker, symbol)
			if next == nil {
				return false, false
			}
			symbol = next
			continue
		}

		return true, true
	}

	return false, false
}

func getNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	textRange := utils.TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[textRange.Pos():textRange.End()]
}

func stripLeadingTypeKeyword(specifierText string) string {
	if !strings.HasPrefix(specifierText, "type") {
		return specifierText
	}

	if len(specifierText) == 4 {
		return ""
	}

	next := rune(specifierText[4])
	if !unicode.IsSpace(next) && next != '/' {
		return specifierText
	}

	return strings.TrimLeft(specifierText[4:], " \t\r\n")
}

func getExportSpecifierText(sourceFile *ast.SourceFile, specifierNode *ast.Node) string {
	specifier := specifierNode.AsExportSpecifier()
	text := strings.TrimSpace(getNodeText(sourceFile, specifierNode))
	if specifier.IsTypeOnly {
		return stripLeadingTypeKeyword(text)
	}

	local := specifier.Name().AsNode()
	if specifier.PropertyName != nil {
		local = specifier.PropertyName.AsNode()
	}
	exported := specifier.Name().AsNode()

	localText := strings.TrimSpace(getNodeText(sourceFile, local))
	exportedText := strings.TrimSpace(getNodeText(sourceFile, exported))
	if localText == exportedText {
		return localText
	}

	return localText + " as " + exportedText
}

func joinWordList(words []string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	case 2:
		return words[0] + " and " + words[1]
	default:
		return strings.Join(words[:len(words)-1], ", ") + ", and " + words[len(words)-1]
	}
}

func buildNamedExportStatement(typeOnly bool, specifiers []string, source string) string {
	statement := "export "
	if typeOnly {
		statement += "type "
	}
	statement += "{ " + strings.Join(specifiers, ", ") + " }"
	if source != "" {
		statement += " from " + source
	}
	statement += ";"
	return statement
}

var ConsistentTypeExportsRule = rule.Rule{
	Name: "consistent-type-exports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[ConsistentTypeExportsOptions](options, "consistent-type-exports")

		checkStarExport := func(node *ast.Node) {
			exportDecl := node.AsExportDeclaration()
			if exportDecl.IsTypeOnly || exportDecl.ModuleSpecifier == nil {
				return
			}
			if exportDecl.ExportClause != nil && !ast.IsNamespaceExport(exportDecl.ExportClause.AsNode()) {
				return
			}

			moduleSymbol := ctx.TypeChecker.GetSymbolAtLocation(exportDecl.ModuleSpecifier.AsNode())
			if moduleSymbol == nil {
				return
			}
			sourceFileType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, moduleSymbol)
			if sourceFileType == nil {
				return
			}

			isThereAnyExportedValue := false

			for _, propertyTypeSymbol := range checker.Checker_getPropertiesOfType(ctx.TypeChecker, sourceFileType) {
				if checker.Checker_getPropertyOfType(ctx.TypeChecker, sourceFileType, propertyTypeSymbol.Name) != nil {
					isThereAnyExportedValue = true
					break
				}
			}
			if isThereAnyExportedValue {
				return
			}

			ctx.ReportNodeWithFixes(node, buildTypeOverValueMessage(), func() []rule.RuleFix {
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Pos())
				for {
					if s.Token() == ast.KindAsteriskToken {
						tokenRange := s.TokenRange()
						return []rule.RuleFix{
							rule.RuleFixReplaceRange(tokenRange.WithEnd(tokenRange.Pos()), "type "),
						}
					}
					if s.Token() == ast.KindEndOfFile || s.TokenRange().Pos() >= node.End() {
						break
					}
					s.Scan()
				}
				return nil
			})
		}

		analyzeNamedExport := func(node *ast.Node) *analyzedNamedExport {
			exportDecl := node.AsExportDeclaration()
			if exportDecl.ExportClause == nil || !ast.IsNamedExports(exportDecl.ExportClause.AsNode()) {
				return nil
			}
			if exportDecl.IsTypeOnly {
				return nil
			}

			report := &analyzedNamedExport{
				node:           node,
				typeTexts:      make([]string, 0, 2),
				typeBasedNodes: make([]*ast.Node, 0, 2),
				typeBasedNames: make([]string, 0, 2),
				valueTexts:     make([]string, 0, 2),
			}
			if exportDecl.ModuleSpecifier != nil {
				report.moduleSource = strings.TrimSpace(getNodeText(ctx.SourceFile, exportDecl.ModuleSpecifier.AsNode()))
			}

			for _, specifierNode := range exportDecl.ExportClause.AsNamedExports().Elements.Nodes {
				specifier := specifierNode.AsExportSpecifier()
				specifierText := getExportSpecifierText(ctx.SourceFile, specifierNode)

				if specifier.IsTypeOnly {
					report.typeTexts = append(report.typeTexts, specifierText)
					continue
				}

				nameNode := specifier.Name().AsNode()
				if specifier.PropertyName != nil {
					nameNode = specifier.PropertyName.AsNode()
				}
				symbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
				isType, resolved := isSymbolTypeBased(ctx.TypeChecker, symbol)
				if !resolved {
					continue
				}

				if isType {
					report.typeTexts = append(report.typeTexts, specifierText)
					report.typeBasedNodes = append(report.typeBasedNodes, specifierNode)
					report.typeBasedNames = append(report.typeBasedNames, specifierText)
				} else {
					report.valueTexts = append(report.valueTexts, specifierText)
				}
			}

			return report
		}

		checkNamedExport := func(report *analyzedNamedExport) {
			if report == nil || len(report.typeBasedNodes) == 0 {
				return
			}

			if len(report.valueTexts) == 0 {
				ctx.ReportNodeWithFixes(report.node, buildTypeOverValueMessage(), func() []rule.RuleFix {
					return []rule.RuleFix{
						rule.RuleFixReplace(
							ctx.SourceFile,
							report.node,
							buildNamedExportStatement(true, report.typeTexts, report.moduleSource),
						),
					}
				})
				return
			}

			msg := buildMultipleExportsAreTypesMessage(joinWordList(report.typeBasedNames))
			if len(report.typeBasedNames) == 1 {
				msg = buildSingleExportIsTypeMessage(report.typeBasedNames[0])
			}

			ctx.ReportNodeWithFixes(report.node, msg, func() []rule.RuleFix {
				if opts.FixMixedExportsWithInlineTypeSpecifier {
					fixes := make([]rule.RuleFix, 0, len(report.typeBasedNodes))
					for _, specifierNode := range report.typeBasedNodes {
						fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, specifierNode, "type "))
					}
					return fixes
				}

				replacement := buildNamedExportStatement(true, report.typeTexts, report.moduleSource) + "\n" +
					buildNamedExportStatement(false, report.valueTexts, report.moduleSource)
				return []rule.RuleFix{
					rule.RuleFixReplace(ctx.SourceFile, report.node, replacement),
				}
			})
		}

		return rule.RuleListeners{
			ast.KindExportDeclaration: func(node *ast.Node) {
				exportDecl := node.AsExportDeclaration()
				if exportDecl.ExportClause == nil || ast.IsNamespaceExport(exportDecl.ExportClause.AsNode()) {
					checkStarExport(node)
					return
				}
				if ast.IsNamedExports(exportDecl.ExportClause.AsNode()) {
					checkNamedExport(analyzeNamedExport(node))
				}
			},
		}
	},
}
