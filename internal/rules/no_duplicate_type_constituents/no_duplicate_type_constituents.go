package no_duplicate_type_constituents

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildDuplicateMessage(unionOrIntersection unionOrIntersection, previous string) rule.RuleMessage {
	var msg string
	if unionOrIntersection == unionOrIntersection_Intersection {
		msg = "Intersection"
	} else {
		msg = "Union"
	}
	return rule.RuleMessage{
		Id:          "duplicate",
		Description: fmt.Sprintf("%v type constituent is duplicated with %v.", msg, previous),
	}
}
func buildUnnecessaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessary",
		Description: "Explicit undefined is unnecessary on an optional parameter.",
	}
}

type unionOrIntersection uint8

const (
	unionOrIntersection_Union unionOrIntersection = iota
	unionOrIntersection_Intersection
)

type NoDuplicateTypeConstituentsOptions struct {
	IgnoreIntersections bool
	IgnoreUnions        bool
}

var NoDuplicateTypeConstituentsRule = rule.Rule{
	Name: "no-duplicate-type-constituents",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(NoDuplicateTypeConstituentsOptions)
		if !ok {
			opts = NoDuplicateTypeConstituentsOptions{
				IgnoreIntersections: false,
				IgnoreUnions:        false,
			}
		}

		unwindedParentType := func(node *ast.Node, kind ast.Kind) *ast.Node {
			for {
				node = node.Parent
				if node == nil {
					return nil
				}

				if node.Kind == kind {
					return node
				}

				if node.Kind != ast.KindParenthesizedType {
					return nil
				}
			}
		}

		report := func(
			withFix bool,
			unionOrIntersection unionOrIntersection,
			message rule.RuleMessage,
			constituentNode *ast.Node,
		) {
			if !withFix {
				ctx.ReportNode(constituentNode, message)
				return
			}

			ctx.ReportNodeWithFixes(constituentNode, message, func() []rule.RuleFix {
				kind := ast.KindUnionType
				if unionOrIntersection == unionOrIntersection_Intersection {
					kind = ast.KindIntersectionType
				}

				parent := unwindedParentType(constituentNode, kind)
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, parent.Pos())
				foundBefore := false
				prevStart := 0
				bracketBeforeTokens := []core.TextRange{}

				for {
					if s.TokenStart() >= constituentNode.Pos() {
						break
					}
					if s.Token() == ast.KindAmpersandToken || s.Token() == ast.KindBarToken {
						foundBefore = true
						prevStart = s.TokenStart()
						bracketBeforeTokens = bracketBeforeTokens[:0]
					} else if s.Token() == ast.KindOpenParenToken {
						bracketBeforeTokens = append(bracketBeforeTokens, s.TokenRange())
					}
					s.Scan()
				}
				fixes := []rule.RuleFix{
					rule.RuleFixRemoveRange(utils.TrimNodeTextRange(ctx.SourceFile, constituentNode)),
				}
				if foundBefore {
					fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(prevStart, prevStart+1)))
					for _, before := range bracketBeforeTokens {
						fixes = append(fixes, rule.RuleFixRemoveRange(before))
					}
					s.ResetPos(constituentNode.End())
					for range bracketBeforeTokens {
						s.Scan()
						if s.Token() != ast.KindCloseParenToken {
							panic(fmt.Sprintf("expected next scanned token to be ')', got '%v'", s.Token()))
						}
						fixes = append(fixes, rule.RuleFixRemoveRange(s.TokenRange()))
					}
				} else {
					s.ResetPos(constituentNode.End())

					closingParensCount := 0
					for {
						s.Scan()

						if s.TokenStart() >= parent.End() {
							panic("couldn't find '&' or '|' token")
						}

						if s.Token() == ast.KindAmpersandToken || s.Token() == ast.KindBarToken {
							fixes = append(fixes, rule.RuleFixRemoveRange(s.TokenRange()))
							break
						}
						if s.Token() != ast.KindCloseParenToken {
							panic(fmt.Sprintf("expected next scanned token to be ')', got '%v'", s.Token()))
						}
						closingParensCount++
						fixes = append(fixes, rule.RuleFixRemoveRange(s.TokenRange()))
					}

					openingParens := make([]core.TextRange, 0, closingParensCount)
					s.ResetPos(parent.Pos())
					for range closingParensCount {
						s.Scan()
						if s.Token() == ast.KindOpenParenToken {
							if len(openingParens) < closingParensCount {
								openingParens = append(openingParens, s.TokenRange())
							}
						} else {
							openingParens = openingParens[:0]
						}

						if s.TokenStart() == constituentNode.Pos() {
							if len(openingParens) != closingParensCount {
								panic(fmt.Sprintf("expected to find %v opening parens, found only %v", closingParensCount, len(openingParens)))
							}
							break
						}

						for _, openingParenRange := range openingParens {
							fixes = append(fixes, rule.RuleFixRemoveRange(openingParenRange))
						}
					}
				}

				return fixes
			})
		}

		var checkDuplicateRecursively func(
			withFix bool,
			unionOrIntersection unionOrIntersection,
			constituentNode *ast.Node,
			cachedTypeMap map[*checker.Type]*ast.Node,
			forEachNodeType func(t *checker.Type, node *ast.Node),
		) bool
		checkDuplicateRecursively = func(
			withFix bool,
			unionOrIntersection unionOrIntersection,
			constituentNode *ast.Node,
			cachedTypeMap map[*checker.Type]*ast.Node,
			forEachNodeType func(t *checker.Type, node *ast.Node),
		) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(constituentNode)
			if utils.IsIntrinsicErrorType(t) {
				return false
			}

			// TODO(port): isSameAstNode just recursively compares two objects to match { a: 1 } | { a: 1 }
			// uniqueConstituents.find(ele => isSameAstNode(ele, constituentNode)) ??
			duplicatedPreviousNode, duplicatedPrevious := cachedTypeMap[t]

			if duplicatedPrevious {
				report(withFix, unionOrIntersection, buildDuplicateMessage(unionOrIntersection, ctx.SourceFile.Text()[duplicatedPreviousNode.Pos():duplicatedPreviousNode.End()]), constituentNode)
				return true
			}

			if forEachNodeType != nil {
				forEachNodeType(t, constituentNode)
			}
			cachedTypeMap[t] = constituentNode

			var types []*ast.Node
			withoutParens := ast.SkipTypeParentheses(constituentNode)
			if unionOrIntersection == unionOrIntersection_Union && withoutParens.Kind == ast.KindUnionType {
				types = withoutParens.AsUnionTypeNode().Types.Nodes
			} else if unionOrIntersection == unionOrIntersection_Intersection && withoutParens.Kind == ast.KindIntersectionType {
				types = withoutParens.AsIntersectionTypeNode().Types.Nodes
			} else {
				return false
			}

			allPrevRemoved := true
			for i, constituent := range types {
				if !checkDuplicateRecursively(
					i < len(types)-1 || !allPrevRemoved,
					unionOrIntersection,
					constituent,
					cachedTypeMap,
					forEachNodeType,
				) {
					allPrevRemoved = false
				}
			}
			return false
		}

		checkDuplicate := func(
			node *ast.Node,
			forEachNodeType func(
				constituentNodeType *checker.Type,
				constituentNode *ast.Node,
			),
		) {
			cachedTypeMap := map[*checker.Type]*ast.Node{}

			var unionOrIntersection unionOrIntersection
			var types []*ast.Node
			if node.Kind == ast.KindIntersectionType {
				unionOrIntersection = unionOrIntersection_Intersection
				types = node.AsIntersectionTypeNode().Types.Nodes
			} else if node.Kind == ast.KindUnionType {
				unionOrIntersection = unionOrIntersection_Union
				types = node.AsUnionTypeNode().Types.Nodes
			} else {
				panic(fmt.Sprintf("expected union or intersection, got %v", node.Kind))
			}

			for _, t := range types {
				checkDuplicateRecursively(
					true,
					unionOrIntersection,
					t,
					cachedTypeMap,
					forEachNodeType,
				)
			}
		}

		ruleListeners := make(rule.RuleListeners, 2)

		if !opts.IgnoreIntersections {
			ruleListeners[ast.KindIntersectionType] = func(node *ast.Node) {
				if unwindedParentType(node, ast.KindIntersectionType) != nil {
					return
				}
				checkDuplicate(node, nil)
			}
		}

		if !opts.IgnoreUnions {
			ruleListeners[ast.KindUnionType] = func(node *ast.Node) {
				if unwindedParentType(node, ast.KindUnionType) != nil {
					return
				}

				checkDuplicate(node, func(constituentNodeType *checker.Type, constituentNode *ast.Node) {
					if !ast.IsParameter(node.Parent) {
						return
					}
					param := node.Parent.AsParameterDeclaration()
					if param.QuestionToken == nil {
						return
					}
					if utils.IsTypeFlagSet(constituentNodeType, checker.TypeFlagsUndefined) {
						report(true, unionOrIntersection_Union, buildUnnecessaryMessage(), constituentNode)
					}
				})
				return
			}
		}

		return ruleListeners
	},
}
