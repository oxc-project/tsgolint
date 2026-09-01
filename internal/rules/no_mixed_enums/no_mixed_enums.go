package no_mixed_enums

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMixedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mixed",
		Description: "Mixing number and string enums can be confusing.",
	}
}

type allowedType uint8

const (
	allowedTypeNumber allowedType = iota
	allowedTypeString
	allowedTypeUnknown
)

var NoMixedEnumsRule = rule.Rule{
	Name: "no-mixed-enums",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		getMemberType := func(node *ast.Node) allowedType {
			initializer := node.AsEnumMember().Initializer

			if initializer == nil {
				return allowedTypeNumber
			}

			switch initializer.Kind {
			case ast.KindNumericLiteral:
				return allowedTypeNumber
			case ast.KindStringLiteral:
				return allowedTypeString
			default:
				t := ctx.TypeChecker.GetTypeAtLocation(initializer)
				if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike) {
					return allowedTypeString
				}
				if utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLike) {
					return allowedTypeNumber
				}
				return allowedTypeUnknown
			}
		}
		getDesiredTypeForDefinition := func(node *ast.Node) allowedType {
			symbol := ctx.TypeChecker.GetSymbolAtLocation(node.Name())

			declaration := symbol.Declarations[0]

			return getMemberType(declaration.Members()[0])
		}
		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enum := node.AsEnumDeclaration()

				if len(enum.Members.Nodes) == 0 {
					return
				}

				desiredType := getDesiredTypeForDefinition(node)
				if desiredType == allowedTypeUnknown {
					return
				}

				for _, member := range enum.Members.Nodes {
					currentType := getMemberType(member)
					if currentType == allowedTypeUnknown {
						return
					}

					if currentType != desiredType {
						init := member.AsEnumMember().Initializer
						if init == nil {
							init = member
						}
						ctx.ReportNode(init, buildMixedMessage())
						return
					}
				}
			},
		}
	},
}
