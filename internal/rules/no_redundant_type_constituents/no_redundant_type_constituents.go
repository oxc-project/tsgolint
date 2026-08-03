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

type seenTypePart struct {
	part     typeFlagsWithNodeOrType
	typeNode *ast.Node
}

type labeledTypePart struct {
	node     *ast.Node
	typeName string
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
			for node.Kind == ast.KindParenthesizedType {
				node = node.AsParenthesizedTypeNode().Type
			}

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

		typePartsToString := func(typeParts []typeFlagsWithNodeOrType) string {
			return strings.Join(utils.Map(typeParts, func(t typeFlagsWithNodeOrType) string {
				return t.ToString(ctx.TypeChecker)
			}), " | ")
		}
		renderTypeParts := func(typeParts []typeFlagsWithNodeOrType) string {
			return strings.Join(utils.Map(typeParts, func(typePart typeFlagsWithNodeOrType) string {
				if typePart.node != nil {
					return ctx.TypeChecker.TypeToString(ctx.TypeChecker.GetTypeAtLocation(typePart.node))
				}
				return ctx.TypeChecker.TypeToString(typePart.t)
			}), " | ")
		}

		typeNodeToString := func(typeNode *ast.Node) string {
			if typeNode == nil {
				return ""
			}
			return typePartsToString(getTypeNodeTypePartFlags(typeNode))
		}

		firstOtherTypeNode := func(typeNodes []*ast.Node, typeNode *ast.Node) *ast.Node {
			for _, candidate := range typeNodes {
				if candidate != typeNode {
					return candidate
				}
			}
			return nil
		}

		renderTypeNode := func(typeNode *ast.Node, fallback string) string {
			if typeNode == nil {
				return fallback
			}
			rendered := renderTypeParts(getTypeNodeTypePartFlags(typeNode))
			if rendered == "" {
				return fallback
			}
			return rendered
		}

		reportRelations := func(
			message rule.RuleMessage,
			primaryNode *ast.Node,
			redundantParts []labeledTypePart,
			overridingParts []labeledTypePart,
		) {
			// Some error and aliased types do not have a distinct local syntax node for
			// both sides of the relationship. Keep the diagnostic useful in those
			// cases, but only attach labels to ranges that are actually available.
			if primaryNode == nil && len(redundantParts) > 0 {
				primaryNode = redundantParts[0].node
			}
			if primaryNode == nil && len(overridingParts) > 0 {
				primaryNode = overridingParts[0].node
			}
			if primaryNode == nil {
				return
			}

			labels := make([]rule.RuleLabeledRange, 0, len(redundantParts)+len(overridingParts))
			for _, part := range redundantParts {
				if part.node == nil {
					continue
				}
				labels = append(labels, rule.RuleLabeledRange{
					Label: fmt.Sprintf("Redundant type: `%s`", part.typeName),
					Range: utils.TrimNodeTextRange(ctx.SourceFile, part.node),
				})
			}
			for _, part := range overridingParts {
				if part.node == nil {
					continue
				}
				labels = append(labels, rule.RuleLabeledRange{
					Label: fmt.Sprintf("Overriding type: `%s`", part.typeName),
					Range: utils.TrimNodeTextRange(ctx.SourceFile, part.node),
				})
			}

			ctx.ReportDiagnostic(rule.RuleDiagnostic{
				Range:         utils.TrimNodeTextRange(ctx.SourceFile, primaryNode),
				Message:       message,
				LabeledRanges: labels,
			})
		}
		reportRelation := func(
			message rule.RuleMessage,
			redundantNode *ast.Node,
			redundantType string,
			overridingNode *ast.Node,
			overridingType string,
		) {
			reportRelations(
				message,
				redundantNode,
				[]labeledTypePart{{node: redundantNode, typeName: redundantType}},
				[]labeledTypePart{{node: overridingNode, typeName: overridingType}},
			)
		}

		checkIntersectionBottomAndTopTypes := func(typePart typeFlagsWithNodeOrType, typeNode *ast.Node, typeNodes []*ast.Node) bool {
			var message rule.RuleMessage
			var redundantNode, overridingNode *ast.Node
			var redundantType, overridingType string

			switch typePart.flags {
			case checker.TypeFlagsAny:
				typeName := typePart.ToString(ctx.TypeChecker)
				if typeName == "any" {
					message = buildOverridesMessage(typeName, "intersection")
				} else {
					message = buildErrorTypeOverridesMessage(typeName, "intersection")
				}
				redundantNode = firstOtherTypeNode(typeNodes, typeNode)
				redundantType = renderTypeNode(redundantNode, typeNodeToString(redundantNode))
				overridingNode = typeNode
				overridingType = renderTypeNode(overridingNode, typeName)
			case checker.TypeFlagsNever:
				typeName := typePart.ToString(ctx.TypeChecker)
				message = buildOverridesMessage(typeName, "intersection")
				redundantNode = firstOtherTypeNode(typeNodes, typeNode)
				redundantType = renderTypeNode(redundantNode, typeNodeToString(redundantNode))
				overridingNode = typeNode
				overridingType = renderTypeNode(overridingNode, typeName)
			case checker.TypeFlagsUnknown:
				redundantNode = typeNode
				redundantType = renderTypeNode(redundantNode, typePart.ToString(ctx.TypeChecker))
				overridingNode = firstOtherTypeNode(typeNodes, typeNode)
				overridingType = renderTypeNode(overridingNode, typeNodeToString(overridingNode))
				message = buildOverriddenMessage(redundantType, "intersection")
			default:
				return false
			}

			reportRelation(message, redundantNode, redundantType, overridingNode, overridingType)
			return true
		}

		return rule.RuleListeners{
			ast.KindIntersectionType: func(node *ast.Node) {
				seenBigIntLiteralTypes := []seenTypePart{}
				seenBooleanLiteralTypes := []seenTypePart{}
				seenNumberLiteralTypes := []seenTypePart{}
				seenStringLiteralTypes := []seenTypePart{}

				seenBigIntPrimitiveTypes := []*ast.Node{}
				seenBooleanPrimitiveTypes := []*ast.Node{}
				seenNumberPrimitiveTypes := []*ast.Node{}
				seenStringPrimitiveTypes := []*ast.Node{}

				seenUnionTypes := []seenUnionPart{}

				typeNodes := node.AsIntersectionTypeNode().Types.Nodes
				for _, typeNode := range typeNodes {
					typePartFlags := getTypeNodeTypePartFlags(typeNode)

					// if any typeNode is TSTypeReference and typePartFlags have more than 1 element, than the referenced type is definitely a union.
					if len(typePartFlags) >= 2 {
						seenUnionTypes = append(seenUnionTypes, seenUnionPart{
							typePartFlags,
							typeNode,
						})
					}

					for _, typePart := range typePartFlags {
						if checkIntersectionBottomAndTopTypes(typePart, typeNode, typeNodes) {
							continue
						}

						// unions assignability check doesn't require seen*LiteralTypes, so avoid computing them
						if len(seenUnionTypes) == 0 {
							switch typePart.flags {
							case checker.TypeFlagsBigIntLiteral:
								seenBigIntLiteralTypes = append(seenBigIntLiteralTypes, seenTypePart{typePart, typeNode})
							case checker.TypeFlagsBooleanLiteral:
								seenBooleanLiteralTypes = append(seenBooleanLiteralTypes, seenTypePart{typePart, typeNode})
							case checker.TypeFlagsNumberLiteral:
								seenNumberLiteralTypes = append(seenNumberLiteralTypes, seenTypePart{typePart, typeNode})
							case checker.TypeFlagsStringLiteral, checker.TypeFlagsTemplateLiteral:
								seenStringLiteralTypes = append(seenStringLiteralTypes, seenTypePart{typePart, typeNode})
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
					for _, unionType := range seenUnionTypes {
						var primitiveName string
						var primitiveNode *ast.Node
						for _, typeValue := range unionType.flags {
							switch {
							case typeValue.flags == checker.TypeFlagsBigIntLiteral && len(seenBigIntPrimitiveTypes) > 0:
								primitiveName = "bigint"
								primitiveNode = seenBigIntPrimitiveTypes[0]
							case typeValue.flags == checker.TypeFlagsBooleanLiteral && len(seenBooleanPrimitiveTypes) > 0:
								primitiveName = "boolean"
								primitiveNode = seenBooleanPrimitiveTypes[0]
							case typeValue.flags == checker.TypeFlagsNumberLiteral && len(seenNumberPrimitiveTypes) > 0:
								primitiveName = "number"
								primitiveNode = seenNumberPrimitiveTypes[0]
							case (typeValue.flags == checker.TypeFlagsStringLiteral || typeValue.flags == checker.TypeFlagsTemplateLiteral) && len(seenStringPrimitiveTypes) > 0:
								primitiveName = "string"
								primitiveNode = seenStringPrimitiveTypes[0]
							default:
								primitiveName = ""
								primitiveNode = nil
							}
							if len(primitiveName) == 0 {
								break
							}
						}

						if len(primitiveName) == 0 {
							continue
						}

						typeValuesLiteral := typePartsToString(unionType.flags)
						renderedTypeValuesLiteral := renderTypeParts(unionType.flags)
						reportRelation(
							buildPrimitiveOverriddenMessage(typeValuesLiteral, primitiveName),
							primitiveNode,
							primitiveName,
							unionType.typeNode,
							renderedTypeValuesLiteral,
						)
					}
				}
				if len(seenUnionTypes) > 0 {
					return
				}

				checkLiteralTypeOverridesPrimitive := func(literalTypes []seenTypePart, primitiveTypes []*ast.Node, primitiveName string) {
					if len(literalTypes) == 0 {
						return
					}
					typeValuesLiteral := strings.Join(utils.Map(literalTypes, func(t seenTypePart) string {
						return t.part.ToString(ctx.TypeChecker)
					}), " | ")
					overridingParts := utils.Map(literalTypes, func(t seenTypePart) labeledTypePart {
						return labeledTypePart{
							node:     t.typeNode,
							typeName: renderTypeParts([]typeFlagsWithNodeOrType{t.part}),
						}
					})
					for _, typeNode := range primitiveTypes {
						reportRelations(
							buildPrimitiveOverriddenMessage(typeValuesLiteral, primitiveName),
							typeNode,
							[]labeledTypePart{{node: typeNode, typeName: primitiveName}},
							overridingParts,
						)
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
				seenBigIntPrimitiveTypeNodes := []*ast.Node{}
				seenBooleanPrimitiveTypeNodes := []*ast.Node{}
				seenNumberPrimitiveTypeNodes := []*ast.Node{}
				seenStringPrimitiveTypeNodes := []*ast.Node{}

				typeNodes := node.AsUnionTypeNode().Types.Nodes
				checkUnionBottomAndTopTypes := func(typePart typeFlagsWithNodeOrType, typeNode *ast.Node) bool {
					var message rule.RuleMessage
					var redundantNode, overridingNode *ast.Node
					var redundantType, overridingType string

					switch typePart.flags {
					case checker.TypeFlagsAny:
						typeName := typePart.ToString(ctx.TypeChecker)
						if typeName == "any" {
							message = buildOverridesMessage(typeName, "union")
						} else {
							message = buildErrorTypeOverridesMessage(typeName, "union")
						}
						redundantNode = firstOtherTypeNode(typeNodes, typeNode)
						redundantType = renderTypeNode(redundantNode, typeNodeToString(redundantNode))
						overridingNode = typeNode
						overridingType = renderTypeNode(overridingNode, typeName)
					case checker.TypeFlagsUnknown:
						typeName := typePart.ToString(ctx.TypeChecker)
						message = buildOverridesMessage(typeName, "union")
						redundantNode = firstOtherTypeNode(typeNodes, typeNode)
						redundantType = renderTypeNode(redundantNode, typeNodeToString(redundantNode))
						overridingNode = typeNode
						overridingType = renderTypeNode(overridingNode, typeName)
					case checker.TypeFlagsNever:
						if isNodeInsideReturnType(node) {
							return false
						}
						redundantNode = typeNode
						redundantType = renderTypeNode(redundantNode, "never")
						overridingNode = firstOtherTypeNode(typeNodes, typeNode)
						overridingType = renderTypeNode(overridingNode, typeNodeToString(overridingNode))
						message = buildOverriddenMessage("never", "union")
					default:
						return false
					}

					reportRelation(message, redundantNode, redundantType, overridingNode, overridingType)
					return true
				}

				for _, typeNode := range typeNodes {
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
						// Aliases of primitive types can carry additional checker flags, so use
						// bit tests here rather than requiring an exact flag match. The alias
						// reference is still a precise local range for the overriding label.
						if typePart.flags&checker.TypeFlagsBigInt != 0 {
							seenBigIntPrimitiveTypeNodes = append(seenBigIntPrimitiveTypeNodes, typeNode)
						}
						if typePart.flags&checker.TypeFlagsBoolean != 0 {
							seenBooleanPrimitiveTypeNodes = append(seenBooleanPrimitiveTypeNodes, typeNode)
						}
						if typePart.flags&checker.TypeFlagsNumber != 0 {
							seenNumberPrimitiveTypeNodes = append(seenNumberPrimitiveTypeNodes, typeNode)
						}
						if typePart.flags&checker.TypeFlagsString != 0 {
							seenStringPrimitiveTypeNodes = append(seenStringPrimitiveTypeNodes, typeNode)
						}
					}
				}

				// For each type node that had at least one overridden literal,
				// group those literals by their primitive type,
				// then report each primitive type with all its literals

				checkOverriddenTypes := func(primitiveFlag checker.TypeFlags, overriddenNodes map[*ast.Node][]typeFlagsWithNodeOrType, primitiveNodes []*ast.Node, primitiveName string) {
					if seenPrimitiveTypeFlags&primitiveFlag == 0 {
						return
					}

					for typeNode, typeFlags := range overriddenNodes {
						typeValuesLiteral := strings.Join(utils.Map(typeFlags, func(t typeFlagsWithNodeOrType) string {
							return t.ToString(ctx.TypeChecker)
						}), " | ")
						redundantParts := utils.Map(typeFlags, func(t typeFlagsWithNodeOrType) labeledTypePart {
							redundantNode := t.node
							if redundantNode == nil {
								redundantNode = typeNode
							}
							return labeledTypePart{
								node:     redundantNode,
								typeName: renderTypeParts([]typeFlagsWithNodeOrType{t}),
							}
						})
						var primitiveNode *ast.Node
						if len(primitiveNodes) > 0 {
							primitiveNode = primitiveNodes[0]
						}
						reportRelations(
							buildLiteralOverriddenMessage(typeValuesLiteral, primitiveName),
							typeNode,
							redundantParts,
							[]labeledTypePart{{node: primitiveNode, typeName: primitiveName}},
						)
					}
				}

				checkOverriddenTypes(checker.TypeFlagsBigInt, overriddenBigIntTypeNodes, seenBigIntPrimitiveTypeNodes, "bigint")
				checkOverriddenTypes(checker.TypeFlagsBoolean, overriddenBooleanTypeNodes, seenBooleanPrimitiveTypeNodes, "boolean")
				checkOverriddenTypes(checker.TypeFlagsNumber, overriddenNumberTypeNodes, seenNumberPrimitiveTypeNodes, "number")
				checkOverriddenTypes(checker.TypeFlagsString, overriddenStringTypeNodes, seenStringPrimitiveTypeNodes, "string")
			},
		}
	},
}
