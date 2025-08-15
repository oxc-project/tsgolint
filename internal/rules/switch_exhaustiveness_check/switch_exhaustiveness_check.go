package switch_exhaustiveness_check

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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
		Description: "Switch is not exhaustive",
	}
}

type SwitchExhaustivenessCheckOptions struct {
	AllowDefaultCaseForExhaustiveSwitch *bool
	ConsiderDefaultExhaustiveForUnions  *bool
	DefaultCaseCommentPattern           *string
	RequireDefaultForNonUnion           *bool
}

type SwitchMetadata struct {
	ContainsNonLiteralType bool
	// nil if there is no default case
	DefaultCase               *ast.CaseOrDefaultClause
	MissingLiteralBranchTypes []*checker.Type
	// TODO: add support for fixed (symbolname is used only for fixes)
	// SymbolName string
}

var SwitchExhaustivenessCheckRule = rule.Rule{
	Name: "switch-exhaustiveness-check",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(SwitchExhaustivenessCheckOptions)
		if !ok {
			opts = SwitchExhaustivenessCheckOptions{}
		}
		if opts.AllowDefaultCaseForExhaustiveSwitch == nil {
			opts.AllowDefaultCaseForExhaustiveSwitch = utils.Ref(true)
		}
		if opts.ConsiderDefaultExhaustiveForUnions == nil {
			opts.ConsiderDefaultExhaustiveForUnions = utils.Ref(false)
		}
		if opts.RequireDefaultForNonUnion == nil {
			opts.RequireDefaultForNonUnion = utils.Ref(false)
		}

		isLiteralLikeType := func(t *checker.Type) bool {
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
		doesTypeContainNonLiteralType := func(t *checker.Type) bool {
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

		getSwitchMetadata := func(node *ast.SwitchStatement) *SwitchMetadata {
			cases := node.CaseBlock.AsCaseBlock().Clauses.Nodes
			defaultCaseIndex := slices.IndexFunc(cases, func(clause *ast.Node) bool {
				return clause.Kind == ast.KindDefaultClause
			})
			var defaultCase *ast.CaseOrDefaultClause
			if defaultCaseIndex > -1 {
				defaultCase = cases[defaultCaseIndex].AsCaseOrDefaultClause()
			}

			discriminantType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node.Expression)

			caseTypes := make([]*checker.Type, 0, len(cases))
			for _, c := range cases {
				if c.Kind == ast.KindDefaultClause {
					continue
				}

				caseTypes = append(caseTypes, utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, c.AsCaseOrDefaultClause().Expression))
			}

			containsNonLiteralType := doesTypeContainNonLiteralType(discriminantType)

			missingLiteralBranchTypes := make([]*checker.Type, 0, 10)
			utils.TypeRecurser(discriminantType, func(t *checker.Type) bool {
				if slices.Contains(caseTypes, t) || !isLiteralLikeType(t) {
					return false
				}

				// "missing", "optional" and "undefined" types are different runtime objects,
				// but all of them have TypeFlags.Undefined type flag
				if slices.ContainsFunc(caseTypes, func(t *checker.Type) bool {
					return utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined)
				}) && utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
					return false
				}

				missingLiteralBranchTypes = append(missingLiteralBranchTypes, t)

				return false
			})

			return &SwitchMetadata{
				ContainsNonLiteralType:    containsNonLiteralType,
				DefaultCase:               defaultCase,
				MissingLiteralBranchTypes: missingLiteralBranchTypes,
			}
		}

		checkSwitchExhaustive := func(node *ast.SwitchStatement, switchMetadata *SwitchMetadata) {
			// If considerDefaultExhaustiveForUnions is enabled, the presence of a default case
			// always makes the switch exhaustive.
			if *opts.ConsiderDefaultExhaustiveForUnions && switchMetadata.DefaultCase != nil {
				return
			}

			if len(switchMetadata.MissingLiteralBranchTypes) > 0 {
				// TODO(port): more verbose message
				//   missingBranches: missingLiteralBranchTypes
				// .map(missingType =>
				//   tsutils.isTypeFlagSet(missingType, ts.TypeFlags.ESSymbolLike)
				//     ? `typeof ${missingType.getSymbol()?.escapedName as string}`
				//     : typeToString(missingType),
				// )
				// .join(' | '),

				ctx.ReportNode(node.Expression, buildSwitchIsNotExhaustiveMessage("TODO"))
			}
		}

		checkSwitchUnnecessaryDefaultCase := func(switchMetadata *SwitchMetadata) {
			if *opts.AllowDefaultCaseForExhaustiveSwitch {
				return
			}

			if len(switchMetadata.MissingLiteralBranchTypes) == 0 &&
				switchMetadata.DefaultCase != nil &&
				!switchMetadata.ContainsNonLiteralType {
				ctx.ReportNode(&switchMetadata.DefaultCase.Node, buildDangerousDefaultCaseMessage())
			}
		}
		checkSwitchNoUnionDefaultCase := func(node *ast.SwitchStatement, switchMetadata *SwitchMetadata) {
			if !*opts.RequireDefaultForNonUnion {
				return
			}

			if switchMetadata.ContainsNonLiteralType && switchMetadata.DefaultCase == nil {
				ctx.ReportNode(node.Expression, buildSwitchIsNotExhaustiveMessage("default"))
				// TODO(port): missing suggestion
			}
		}

		return rule.RuleListeners{
			ast.KindSwitchStatement: func(node *ast.Node) {

				stmt := node.AsSwitchStatement()

				metadata := getSwitchMetadata(stmt)
				checkSwitchExhaustive(stmt, metadata)
				checkSwitchUnnecessaryDefaultCase(metadata)
				checkSwitchNoUnionDefaultCase(stmt, metadata)
			},
		}

	},
}
