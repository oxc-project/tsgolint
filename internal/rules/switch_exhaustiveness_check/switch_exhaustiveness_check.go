package switch_exhaustiveness_check

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildAddMissingCasesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addMissingCases",
		Description: "Add branches for missing cases.",
	}
}

func buildDangerousDefaultCaseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "dangerousDefaultCase",
		Description: "The switch statement is exhaustive, so the default case is unnecessary.",
	}
}
func buildSwitchIsNotExhaustiveMessage(missingBranches string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "switchIsNotExhaustive",
		Description: "Switch is not exhaustive. Cases not matched: " + missingBranches,
	}
}

type SwitchMetadata struct {
	ContainsNonLiteralType bool
	// nil if there is no default case
	DefaultCase               *ast.CaseOrDefaultClause
	MissingLiteralBranchTypes []*checker.Type
	SymbolName                string
}

func isLiteralLikeType(t *checker.Type) bool {
	return utils.IsTypeFlagSet(
		t,
		checker.TypeFlagsLiteral|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsUniqueESSymbol,
	)
}

/**
 * For example:
 *
 * - `"foo" | "bar"` is a type with all literal types.
 * - `"foo" | number` is a type that contains non-literal types.
 * - `"foo" & { bar: 1 }` is a type that contains non-literal types.
 *
 * Default cases are never superfluous in switches with non-literal types.
 */
func doesTypeContainNonLiteralType(t *checker.Type) bool {
	return utils.Some(
		utils.UnionTypeParts(t),
		func(t *checker.Type) bool {
			return utils.Every(
				utils.IntersectionTypeParts(t),
				func(t *checker.Type) bool {
					return !isLiteralLikeType(t)
				},
			)
		},
	)
}

func getSwitchMetadata(typeChecker *checker.Checker, node *ast.SwitchStatement) *SwitchMetadata {
	cases := node.CaseBlock.AsCaseBlock().Clauses.Nodes
	defaultCaseIndex := slices.IndexFunc(cases, func(clause *ast.Node) bool {
		return clause.Kind == ast.KindDefaultClause
	})
	var defaultCase *ast.CaseOrDefaultClause
	if defaultCaseIndex > -1 {
		defaultCase = cases[defaultCaseIndex].AsCaseOrDefaultClause()
	}

	discriminantType := utils.GetConstrainedTypeAtLocation(typeChecker, node.Expression)
	symbolName := ""
	if discriminantType.Symbol() != nil {
		symbolName = discriminantType.Symbol().Name
	}

	caseTypeSet := make(map[*checker.Type]struct{}, len(cases))
	hasUndefinedCase := false
	for _, c := range cases {
		if c.Kind == ast.KindDefaultClause {
			continue
		}

		caseType := utils.GetConstrainedTypeAtLocation(typeChecker, c.AsCaseOrDefaultClause().Expression)
		caseTypeSet[caseType] = struct{}{}
		if utils.IsTypeFlagSet(caseType, checker.TypeFlagsUndefined) {
			hasUndefinedCase = true
		}
	}

	containsNonLiteralType := doesTypeContainNonLiteralType(discriminantType)

	missingLiteralBranchTypes := make([]*checker.Type, 0, 10)
	utils.TypeRecurser(discriminantType, func(t *checker.Type) bool {
		if _, ok := caseTypeSet[t]; ok || !isLiteralLikeType(t) {
			return false
		}

		// "missing", "optional" and "undefined" types are different runtime objects,
		// but all of them have TypeFlags.Undefined type flag
		if hasUndefinedCase && utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
			return false
		}

		missingLiteralBranchTypes = append(missingLiteralBranchTypes, t)

		return false
	})

	return &SwitchMetadata{
		ContainsNonLiteralType:    containsNonLiteralType,
		DefaultCase:               defaultCase,
		MissingLiteralBranchTypes: missingLiteralBranchTypes,
		SymbolName:                symbolName,
	}
}

func requiresQuoting(text string) bool {
	return !scanner.IsIdentifierText(text, core.LanguageVariantStandard)
}

func getNodeIndent(sourceFile *ast.SourceFile, node *ast.Node) string {
	trimmed := utils.TrimNodeTextRange(sourceFile, node)
	_, column := scanner.GetECMALineAndCharacterOfPosition(sourceFile, trimmed.Pos())
	if column <= 0 {
		return ""
	}
	return strings.Repeat(" ", column)
}

func buildCaseTest(typeChecker *checker.Checker, missingBranchType *checker.Type, symbolName string) string {
	missingBranchName := ""
	if missingBranchType.Symbol() != nil {
		missingBranchName = missingBranchType.Symbol().Name
	}

	caseTest := ""
	if utils.IsTypeFlagSet(missingBranchType, checker.TypeFlagsESSymbolLike) {
		caseTest = missingBranchName
	} else {
		caseTest = typeChecker.TypeToString(missingBranchType)
	}

	if symbolName != "" && requiresQuoting(missingBranchName) {
		escapedBranchName := strings.NewReplacer(
			"'", "\\'",
			"\n", "\\n",
			"\r", "\\r",
		).Replace(missingBranchName)

		return fmt.Sprintf("%s['%s']", symbolName, escapedBranchName)
	}

	return caseTest
}

func buildMissingCaseLine(typeChecker *checker.Checker, missingBranchType *checker.Type, symbolName string) string {
	if missingBranchType == nil {
		return "default: { throw new Error('default case') }"
	}

	caseTest := buildCaseTest(typeChecker, missingBranchType, symbolName)
	escapedCase := strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(caseTest)
	return fmt.Sprintf("case %s: { throw new Error('Not implemented yet: %s case') }", caseTest, escapedCase)
}

func buildMissingCases(typeChecker *checker.Checker, missingBranchTypes []*checker.Type, symbolName string) []string {
	missingCases := make([]string, 0, len(missingBranchTypes))
	for _, missingBranchType := range missingBranchTypes {
		missingCases = append(missingCases, buildMissingCaseLine(typeChecker, missingBranchType, symbolName))
	}
	return missingCases
}

func applyMissingCases(sourceFile *ast.SourceFile, node *ast.SwitchStatement, defaultCase *ast.CaseOrDefaultClause, missingCases []string) []rule.RuleFix {
	cases := node.CaseBlock.AsCaseBlock().Clauses.Nodes

	var lastCase *ast.Node
	if len(cases) > 0 {
		lastCase = cases[len(cases)-1]
	}

	caseIndent := ""
	if lastCase != nil {
		caseIndent = getNodeIndent(sourceFile, lastCase)
	} else {
		caseIndent = getNodeIndent(sourceFile, &node.Node)
	}

	fixString := ""
	if len(missingCases) > 0 {
		indented := make([]string, 0, len(missingCases))
		for _, code := range missingCases {
			indented = append(indented, caseIndent+code)
		}
		fixString = strings.Join(indented, "\n")
	}

	if lastCase != nil {
		if defaultCase != nil {
			beforeFixBuilder := strings.Builder{}
			for _, code := range missingCases {
				beforeFixBuilder.WriteString(code)
				beforeFixBuilder.WriteByte('\n')
				beforeFixBuilder.WriteString(caseIndent)
			}

			return []rule.RuleFix{
				rule.RuleFixInsertBefore(sourceFile, &defaultCase.Node, beforeFixBuilder.String()),
			}
		}

		return []rule.RuleFix{
			rule.RuleFixInsertAfter(lastCase, "\n"+fixString),
		}
	}

	return []rule.RuleFix{
		rule.RuleFixReplace(
			sourceFile,
			node.CaseBlock,
			strings.Join([]string{"{", fixString, caseIndent + "}"}, "\n"),
		),
	}
}

func fixSwitch(sourceFile *ast.SourceFile, typeChecker *checker.Checker, node *ast.SwitchStatement, missingBranchTypes []*checker.Type, defaultCase *ast.CaseOrDefaultClause, symbolName string) []rule.RuleFix {
	missingCases := buildMissingCases(typeChecker, missingBranchTypes, symbolName)
	return applyMissingCases(sourceFile, node, defaultCase, missingCases)
}

var SwitchExhaustivenessCheckRule = rule.Rule{
	Name: "switch-exhaustiveness-check",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[SwitchExhaustivenessCheckOptions](options, "switch-exhaustiveness-check")

		checkSwitchExhaustive := func(node *ast.SwitchStatement, switchMetadata *SwitchMetadata) {
			// If considerDefaultExhaustiveForUnions is enabled, the presence of a default case
			// always makes the switch exhaustive.
			if opts.ConsiderDefaultExhaustiveForUnions && switchMetadata.DefaultCase != nil {
				return
			}

			if len(switchMetadata.MissingLiteralBranchTypes) > 0 {
				missingBranches := make([]string, 0, len(switchMetadata.MissingLiteralBranchTypes))
				for _, missingType := range switchMetadata.MissingLiteralBranchTypes {
					if utils.IsTypeFlagSet(missingType, checker.TypeFlagsESSymbolLike) && missingType.Symbol() != nil {
						missingBranches = append(missingBranches, "typeof "+missingType.Symbol().Name)
						continue
					}

					missingBranches = append(missingBranches, ctx.TypeChecker.TypeToString(missingType))
				}

				ctx.ReportNodeWithSuggestions(node.Expression, buildSwitchIsNotExhaustiveMessage(strings.Join(missingBranches, " | ")), func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message:  buildAddMissingCasesMessage(),
						FixesArr: fixSwitch(ctx.SourceFile, ctx.TypeChecker, node, switchMetadata.MissingLiteralBranchTypes, switchMetadata.DefaultCase, switchMetadata.SymbolName),
					}}
				})
			}
		}

		checkSwitchUnnecessaryDefaultCase := func(switchMetadata *SwitchMetadata) {
			if opts.AllowDefaultCaseForExhaustiveSwitch {
				return
			}

			if len(switchMetadata.MissingLiteralBranchTypes) == 0 &&
				switchMetadata.DefaultCase != nil &&
				!switchMetadata.ContainsNonLiteralType {
				ctx.ReportNode(&switchMetadata.DefaultCase.Node, buildDangerousDefaultCaseMessage())
			}
		}
		checkSwitchNoUnionDefaultCase := func(node *ast.SwitchStatement, switchMetadata *SwitchMetadata) {
			if !opts.RequireDefaultForNonUnion {
				return
			}

			if switchMetadata.ContainsNonLiteralType && switchMetadata.DefaultCase == nil {
				ctx.ReportNodeWithSuggestions(node.Expression, buildSwitchIsNotExhaustiveMessage("default"), func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message:  buildAddMissingCasesMessage(),
						FixesArr: fixSwitch(ctx.SourceFile, ctx.TypeChecker, node, []*checker.Type{nil}, switchMetadata.DefaultCase, switchMetadata.SymbolName),
					}}
				})
			}
		}

		return rule.RuleListeners{
			ast.KindSwitchStatement: func(node *ast.Node) {

				stmt := node.AsSwitchStatement()

				metadata := getSwitchMetadata(ctx.TypeChecker, stmt)
				checkSwitchExhaustive(stmt, metadata)
				checkSwitchUnnecessaryDefaultCase(metadata)
				checkSwitchNoUnionDefaultCase(stmt, metadata)
			},
		}

	},
}
