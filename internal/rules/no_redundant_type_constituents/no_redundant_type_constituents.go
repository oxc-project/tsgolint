package no_redundant_type_constituents

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildErrorTypeOverridesMessage(typeName, container string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorTypeOverrides",
		Description: fmt.Sprintf("'%v' is an 'error' type that acts as 'any' and overrides all other types in this %v type.", typeName, container),
	}
}
func buildLiteralOverriddenMessage(literal, primitive string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "literalOverridden",
		Description: fmt.Sprintf("%v is overridden by %v in this union type.", literal, primitive),
	}
}
func buildOverriddenMessage(typeName, container string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "overridden",
		Description: fmt.Sprintf("'%v' is overridden by other types in this %v type.", typeName, container),
	}
}
func buildOverridesMessage(typeName, container string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "overrides",
		Description: fmt.Sprintf("'%v' overrides all other types in this %v type.", typeName, container),
	}
}
func buildPrimitiveOverriddenMessage(literal, primitive string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "primitiveOverridden",
		Description: fmt.Sprintf("%v is overridden by the %v in this intersection type.", primitive, literal),
	}
}

func isNodeInsideReturnType(node *ast.Node) bool {
	return ast.IsFunctionLike(node.Parent)
}

type typeFlagsWithNodeOrType struct {
	flags checker.TypeFlags
	// either node or t must be non-nil
	node *ast.Node
	t    *checker.Type
}

type seenUnionPart struct {
	flags    []typeFlagsWithNodeOrType
	typeNode *ast.Node
}

func (t *typeFlagsWithNodeOrType) ToString(typeChecker *checker.Checker) string {
	if t.node != nil {
		switch t.node.Kind {
		case ast.KindAnyKeyword:
			return "any"
		case ast.KindBooleanKeyword:
			return "boolean"
		case ast.KindNeverKeyword:
			return "never"
		case ast.KindNumberKeyword:
			return "number"
		case ast.KindStringKeyword:
			return "string"
		case ast.KindUnknownKeyword:
			return "unknown"
		case ast.KindLiteralType:
			literal := t.node.AsLiteralTypeNode().Literal
			switch literal.Kind {
			case ast.KindTemplateLiteralType, ast.KindNoSubstitutionTemplateLiteral:
				return "template literal type"
			case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral:
				return literal.Text()
			}
		}
		return "literal type"
	}

	if utils.IsTypeFlagSet(t.t, checker.TypeFlagsStringLiteral) {
		return fmt.Sprintf("%q", typeChecker.TypeToString(t.t))
	}

	return typeChecker.TypeToString(t.t)
}

var NoRedundantTypeConstituentsRule = rule.Rule{
	Name: "no-redundant-type-constituents",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var getTypeNodeTypePartFlags func(node *ast.Node) []typeFlagsWithNodeOrType
		getTypeNodeTypePartFlags = func(node *ast.Node) []typeFlagsWithNodeOrType {
			node = ast.SkipParentheses(node)

			flags := checker.TypeFlagsNone
			switch node.Kind {
			case ast.KindAnyKeyword:
				flags = checker.TypeFlagsAny
			case ast.KindBigIntKeyword:
				flags = checker.TypeFlagsBigInt
			case ast.KindBooleanKeyword:
				flags = checker.TypeFlagsBoolean
			case ast.KindNeverKeyword:
				flags = checker.TypeFlagsNever
			case ast.KindNumberKeyword:
				flags = checker.TypeFlagsNumber
			case ast.KindStringKeyword:
				flags = checker.TypeFlagsString
			case ast.KindUnknownKeyword:
				flags = checker.TypeFlagsUnknown
			}

			if flags != checker.TypeFlagsNone {
				return []typeFlagsWithNodeOrType{{
					flags: flags,
					node:  node,
				}}
			}

			if ast.IsLiteralTypeNode(node) {
				switch node.AsLiteralTypeNode().Literal.Kind {
				case ast.KindBigIntLiteral:
					flags = checker.TypeFlagsBigIntLiteral
				case ast.KindTrueKeyword, ast.KindFalseKeyword:
					flags = checker.TypeFlagsBooleanLiteral
				case ast.KindNumericLiteral:
					flags = checker.TypeFlagsNumberLiteral
				case ast.KindStringLiteral:
					flags = checker.TypeFlagsStringLiteral
				}

				if flags != checker.TypeFlagsNone {
					return []typeFlagsWithNodeOrType{{
						flags: flags,
						node:  node,
					}}
				}
			}

			if node.Kind == ast.KindUnionType {
				var result []typeFlagsWithNodeOrType
				for _, subArray := range node.AsUnionTypeNode().Types.Nodes {
					result = append(result, getTypeNodeTypePartFlags(subArray)...)
				}
				return result
			}

			t := ctx.TypeChecker.GetTypeAtLocation(node)

			var typeParts []*checker.Type
			if t == checker.Checker_booleanType(ctx.TypeChecker) {
				typeParts = []*checker.Type{t}
			} else {
				typeParts = utils.UnionTypeParts(t)
			}

			res := make([]typeFlagsWithNodeOrType, len(typeParts))
			for i, part := range typeParts {
				res[i] = typeFlagsWithNodeOrType{
					flags: checker.Type_flags(part),
					t:     part,
				}
			}
			return res
		}

		checkIntersectionBottomAndTopTypes := func(typePart typeFlagsWithNodeOrType, typeNode *ast.Node) bool {
			var message rule.RuleMessage

			switch typePart.flags {
			case checker.TypeFlagsAny:
				typeName := typePart.ToString(ctx.TypeChecker)
				if typeName == "any" {
					message = buildOverridesMessage(typeName, "intersection")
				} else {
					message = buildErrorTypeOverridesMessage(typeName, "intersection")
				}
			case checker.TypeFlagsNever:
				message = buildOverridesMessage(typePart.ToString(ctx.TypeChecker), "intersection")
			case checker.TypeFlagsUnknown:
				message = buildOverriddenMessage(typePart.ToString(ctx.TypeChecker), "intersection")
			default:
				return false
			}

			ctx.ReportNode(typeNode, message)
			return true
		}

		return rule.RuleListeners{
			ast.KindIntersectionType: func(node *ast.Node) {
				seenBigIntLiteralTypes := []typeFlagsWithNodeOrType{}
				seenBooleanLiteralTypes := []typeFlagsWithNodeOrType{}
				seenNumberLiteralTypes := []typeFlagsWithNodeOrType{}
				seenStringLiteralTypes := []typeFlagsWithNodeOrType{}

				seenBigIntPrimitiveTypes := []*ast.Node{}
				seenBooleanPrimitiveTypes := []*ast.Node{}
				seenNumberPrimitiveTypes := []*ast.Node{}
				seenStringPrimitiveTypes := []*ast.Node{}

				seenUnionTypes := []seenUnionPart{}

				for _, typeNode := range node.AsIntersectionTypeNode().Types.Nodes {
					typePartFlags := getTypeNodeTypePartFlags(typeNode)

					// if any typeNode is TSTypeReference and typePartFlags have more than 1 element, than the referenced type is definitely a union.
					if len(typePartFlags) >= 2 {
						seenUnionTypes = append(seenUnionTypes, seenUnionPart{
							typePartFlags,
							typeNode,
						})
					}

					for _, typePart := range typePartFlags {
						if checkIntersectionBottomAndTopTypes(typePart, typeNode) {
							continue
						}

						// unions assignability check doesn't require seen*LiteralTypes, so avoid computing them
						if len(seenUnionTypes) == 0 {
							switch typePart.flags {
							case checker.TypeFlagsBigIntLiteral:
								seenBigIntLiteralTypes = append(seenBigIntLiteralTypes, typePart)
							case checker.TypeFlagsBooleanLiteral:
								seenBooleanLiteralTypes = append(seenBooleanLiteralTypes, typePart)
							case checker.TypeFlagsNumberLiteral:
								seenNumberLiteralTypes = append(seenNumberLiteralTypes, typePart)
							case checker.TypeFlagsStringLiteral, checker.TypeFlagsTemplateLiteral:
								seenStringLiteralTypes = append(seenStringLiteralTypes, typePart)
							}
						}

						switch typePart.flags {
						case checker.TypeFlagsBigInt:
							seenBigIntPrimitiveTypes = append(seenBigIntPrimitiveTypes, typeNode)
						case checker.TypeFlagsBoolean:
							seenBooleanPrimitiveTypes = append(seenBooleanPrimitiveTypes, typeNode)
						case checker.TypeFlagsNumber:
							seenNumberPrimitiveTypes = append(seenNumberPrimitiveTypes, typeNode)
						case checker.TypeFlagsString:
							seenStringPrimitiveTypes = append(seenStringPrimitiveTypes, typeNode)
						}
					}
				}

				/**
				 * @example
				 * ```ts
				 * type F = "a"|2|"b";
				 * type I = F & string;
				 * ```
				 * This function checks if all the union members of `F` are assignable to the other member of `I`. If every member is assignable, then its reported else not.
				 */
				if len(seenUnionTypes) > 0 && (len(seenBigIntPrimitiveTypes) > 0 || len(seenBooleanPrimitiveTypes) > 0 || len(seenNumberPrimitiveTypes) > 0 || len(seenStringPrimitiveTypes) > 0) {
					var typeValuesLiteral string

					for _, unionType := range seenUnionTypes {
						var primitiveName string
						for _, typeValue := range unionType.flags {
							switch {
							case typeValue.flags == checker.TypeFlagsBigIntLiteral && len(seenBigIntPrimitiveTypes) > 0:
								primitiveName = "bigint"
							case typeValue.flags == checker.TypeFlagsBooleanLiteral && len(seenBooleanPrimitiveTypes) > 0:
								primitiveName = "boolean"
							case typeValue.flags == checker.TypeFlagsNumberLiteral && len(seenNumberPrimitiveTypes) > 0:
								primitiveName = "number"
							case (typeValue.flags == checker.TypeFlagsStringLiteral || typeValue.flags == checker.TypeFlagsTemplateLiteral) && len(seenStringPrimitiveTypes) > 0:
								primitiveName = "string"
							default:
								primitiveName = ""
							}
							if len(primitiveName) == 0 {
								break
							}
						}

						if len(primitiveName) == 0 {
							continue
						}

						if len(typeValuesLiteral) == 0 {
							typeValuesLiteral = strings.Join(utils.Map(unionType.flags, func(t typeFlagsWithNodeOrType) string {
								return t.ToString(ctx.TypeChecker)
							}), " | ")
						}
						ctx.ReportNode(unionType.typeNode, buildPrimitiveOverriddenMessage(primitiveName, typeValuesLiteral))
					}
				}
				if len(seenUnionTypes) > 0 {
					return
				}

				checkLiteralTypeOverridesPrimitive := func(literalTypes []typeFlagsWithNodeOrType, primitiveTypes []*ast.Node, primitiveName string) {
					if len(literalTypes) == 0 {
						return
					}
					typeValuesLiteral := strings.Join(utils.Map(literalTypes, func(t typeFlagsWithNodeOrType) string {
						return t.ToString(ctx.TypeChecker)
					}), " | ")
					for _, typeNode := range primitiveTypes {
						ctx.ReportNode(typeNode, buildPrimitiveOverriddenMessage(typeValuesLiteral, primitiveName))
					}
				}

				// For each primitive type of all the seen primitive types,
				// if there was a literal type seen that overrides it,
				// report each of the primitive type's type nodes
				checkLiteralTypeOverridesPrimitive(seenBigIntLiteralTypes, seenBigIntPrimitiveTypes, "bigint")
				checkLiteralTypeOverridesPrimitive(seenBooleanLiteralTypes, seenBooleanPrimitiveTypes, "boolean")
				checkLiteralTypeOverridesPrimitive(seenNumberLiteralTypes, seenNumberPrimitiveTypes, "number")
				checkLiteralTypeOverridesPrimitive(seenStringLiteralTypes, seenStringPrimitiveTypes, "string")
			},
			ast.KindUnionType: func(node *ast.Node) {
				overriddenBigIntTypeNodes := map[*ast.Node][]typeFlagsWithNodeOrType{}
				overriddenBooleanTypeNodes := map[*ast.Node][]typeFlagsWithNodeOrType{}
				overriddenNumberTypeNodes := map[*ast.Node][]typeFlagsWithNodeOrType{}
				overriddenStringTypeNodes := map[*ast.Node][]typeFlagsWithNodeOrType{}

				seenPrimitiveTypeFlags := checker.TypeFlagsNone

				checkUnionBottomAndTopTypes := func(typePart typeFlagsWithNodeOrType, typeNode *ast.Node) bool {
					var message rule.RuleMessage

					switch typePart.flags {
					case checker.TypeFlagsAny:
						typeName := typePart.ToString(ctx.TypeChecker)
						if typeName == "any" {
							message = buildOverridesMessage(typeName, "union")
						} else {
							message = buildErrorTypeOverridesMessage(typeName, "union")
						}
					case checker.TypeFlagsUnknown:
						message = buildOverridesMessage(typePart.ToString(ctx.TypeChecker), "union")
					case checker.TypeFlagsNever:
						if isNodeInsideReturnType(node) {
							return false
						}
						message = buildOverriddenMessage("never", "union")
					default:
						return false
					}

					ctx.ReportNode(typeNode, message)
					return true
				}

				for _, typeNode := range node.AsUnionTypeNode().Types.Nodes {
					typePartFlags := getTypeNodeTypePartFlags(typeNode)

					for _, typePart := range typePartFlags {
						if checkUnionBottomAndTopTypes(typePart, typeNode) {
							continue
						}

						// For each primitive type of all the seen literal types,
						// if there was a primitive type seen that overrides it,
						// upsert the literal text and primitive type under the backing type node
						switch typePart.flags {
						case checker.TypeFlagsBigIntLiteral:
							overriddenBigIntTypeNodes[typeNode] = append(overriddenBigIntTypeNodes[typeNode], typePart)
						case checker.TypeFlagsBooleanLiteral:
							overriddenBooleanTypeNodes[typeNode] = append(overriddenBooleanTypeNodes[typeNode], typePart)
						case checker.TypeFlagsNumberLiteral:
							overriddenNumberTypeNodes[typeNode] = append(overriddenNumberTypeNodes[typeNode], typePart)
						case checker.TypeFlagsStringLiteral, checker.TypeFlagsTemplateLiteral:
							overriddenStringTypeNodes[typeNode] = append(overriddenStringTypeNodes[typeNode], typePart)
						}

						seenPrimitiveTypeFlags |= typePart.flags & (checker.TypeFlagsBigInt | checker.TypeFlagsBoolean | checker.TypeFlagsNumber | checker.TypeFlagsString)
					}
				}

				// For each type node that had at least one overridden literal,
				// group those literals by their primitive type,
				// then report each primitive type with all its literals

				checkOverriddenTypes := func(primitiveFlag checker.TypeFlags, overriddenNodes map[*ast.Node][]typeFlagsWithNodeOrType, primitiveName string) {
					if seenPrimitiveTypeFlags&primitiveFlag == 0 {
						return
					}

					for typeNode, typeFlags := range overriddenNodes {
						ctx.ReportNode(typeNode, buildLiteralOverriddenMessage(strings.Join(utils.Map(typeFlags, func(t typeFlagsWithNodeOrType) string {
							return t.ToString(ctx.TypeChecker)
						}), " | "), primitiveName))

					}
				}

				checkOverriddenTypes(checker.TypeFlagsBigInt, overriddenBigIntTypeNodes, "bigint")
				checkOverriddenTypes(checker.TypeFlagsBoolean, overriddenBooleanTypeNodes, "boolean")
				checkOverriddenTypes(checker.TypeFlagsNumber, overriddenNumberTypeNodes, "number")
				checkOverriddenTypes(checker.TypeFlagsString, overriddenStringTypeNodes, "string")
			},
		}
	},
}
