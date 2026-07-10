package no_unsafe_argument

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildUnsafeArgumentMessage(sender string, receiver string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArgument",
		Description: fmt.Sprintf("Unsafe argument of type %v assigned to a parameter of type %v.", sender, receiver),
	}
}
func buildUnsafeArraySpreadMessage(sender string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArraySpread",
		Description: fmt.Sprintf("Unsafe spread of an %v array type.", sender),
	}
}
func buildUnsafeSpreadMessage(sender string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeSpread",
		Description: fmt.Sprintf("Unsafe spread of an %v type.", sender),
	}
}
func buildUnsafeTupleSpreadMessage(sender string, receiver string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeTupleSpread",
		Description: fmt.Sprintf("Unsafe spread of a tuple type. The argument is %v and is assigned to a parameter of type %v.", sender, receiver),
	}
}

type restTypeKind uint32

const (
	restTypeKindArray restTypeKind = iota
	restTypeKindTuple
	restTypeKindOther
)

type restType struct {
	Index         int
	Kind          restTypeKind
	Type          *checker.Type
	TypeArguments []*checker.Type
}

func newFunctionSignature(
	typeChecker *checker.Checker,
	node *ast.Node,
) *functionSignature {
	signature := checker.Checker_getResolvedSignature(typeChecker, node, nil, checker.CheckModeNormal)
	if signature == nil {
		return nil
	}

	parameters := checker.Signature_parameters(signature)
	var restParam *ast.Symbol

	for i, param := range parameters {
		if param.Declarations != nil && len(param.Declarations) != 0 {
			if utils.IsRestParameterDeclaration(param.Declarations[0]) {
				// is a rest param
				restParam = param
				parameters = parameters[:i]
				break
			}
		}
	}

	return &functionSignature{
		typeChecker: typeChecker,
		node:        node,
		parameters:  parameters,
		paramTypes:  make([]*checker.Type, len(parameters)),
		restParam:   restParam,
	}
}

type functionSignature struct {
	hasConsumedArguments bool
	parameterTypeIndex   int

	typeChecker *checker.Checker
	node        *ast.Node
	// parameters holds the non-rest parameters; paramTypes caches their types,
	// resolved lazily so that skipped argument positions never request them.
	parameters []*ast.Symbol
	paramTypes []*checker.Type
	restParam  *ast.Symbol
	restType   *restType
}

func (s *functionSignature) consumeRemainingArguments() {
	s.hasConsumedArguments = true
}

// skipParameter advances past the parameter position of an argument that
// doesn't need checking, without resolving the parameter's type.
func (s *functionSignature) skipParameter() {
	s.parameterTypeIndex += 1
}

func (s *functionSignature) getRestType() *restType {
	if s.restType != nil {
		return s.restType
	}
	restT := restType{Index: len(s.parameters), Kind: restTypeKindOther}
	if s.restParam != nil {
		t := s.typeChecker.GetTypeOfSymbolAtLocation(s.restParam, s.node)
		if checker.Checker_isArrayType(s.typeChecker, t) {
			restT.Kind = restTypeKindArray
			restT.Type = checker.Checker_getTypeArguments(s.typeChecker, t)[0]
		} else if checker.IsTupleType(t) {
			restT.Kind = restTypeKindTuple
			restT.TypeArguments = checker.Checker_getTypeArguments(s.typeChecker, t)
		} else {
			restT.Type = t
		}
	}
	s.restType = &restT
	return s.restType
}

func (s *functionSignature) getNextParameterType() *checker.Type {
	index := s.parameterTypeIndex
	s.parameterTypeIndex += 1

	if index >= len(s.paramTypes) || s.hasConsumedArguments {
		restType := s.getRestType()

		switch restType.Kind {
		case restTypeKindTuple:
			typeArguments := restType.TypeArguments
			if len(typeArguments) == 0 {
				return nil
			}
			if s.hasConsumedArguments {
				// all types consumed by a rest - just assume it's the last type
				// there is one edge case where this is wrong, but we ignore it because
				// it's rare and really complicated to handle
				// eg: function foo(...a: [number, ...string[], number])
				return typeArguments[len(typeArguments)-1]
			}

			typeIndex := index - restType.Index
			if typeIndex >= len(typeArguments) {
				return typeArguments[len(typeArguments)-1]
			}

			return typeArguments[typeIndex]
		case restTypeKindArray, restTypeKindOther:
			return restType.Type
		}
	}
	if s.paramTypes[index] == nil {
		s.paramTypes[index] = s.typeChecker.GetTypeOfSymbolAtLocation(s.parameters[index], s.node)
	}
	return s.paramTypes[index]
}

// argumentCanBeUnsafe reports whether an argument expression could have a type
// that IsUnsafeAssignment flags: `any`, or a generic type reference with unsafe
// type arguments. Literals and template strings have literal/primitive types,
// and function/object-literal expressions have anonymous object types — never
// `any` and never a type reference — so resolving their type can be skipped.
// Array literals CAN be unsafe (their type is the generic Array<T> reference,
// e.g. `foo([anyValue])`), so they are not skipped.
func argumentCanBeUnsafe(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral,
		ast.KindArrowFunction,
		ast.KindFunctionExpression:
		return false
	case ast.KindObjectLiteralExpression:
		// Spreading an `any` value makes the whole object literal `any`
		// (e.g. `foo({ ...anyValue })`), so only spread-free literals are safe.
		return slices.ContainsFunc(node.AsObjectLiteralExpression().Properties.Nodes, func(property *ast.Node) bool {
			return property.Kind == ast.KindSpreadAssignment
		})
	default:
		return true
	}
}

var NoUnsafeArgumentRule = rule.Rule{
	Name: "no-unsafe-argument",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		describeType := func(t *checker.Type) string {
			if utils.IsIntrinsicErrorType(t) {
				return "error typed"
			}

			return ctx.TypeChecker.TypeToString(t)
		}

		describeTypeForSpread := func(t *checker.Type) string {
			if checker.Checker_isArrayType(ctx.TypeChecker, t) && utils.IsIntrinsicErrorType(checker.Checker_getTypeArguments(ctx.TypeChecker, t)[0]) {
				return "error"
			}

			return describeType(t)
		}

		describeTypeForTuple := func(t *checker.Type) string {
			if utils.IsIntrinsicErrorType(t) {
				return "error typed"
			}

			return "of type " + ctx.TypeChecker.TypeToString(t)
		}

		checkUnsafeArguments := func(
			args []*ast.Node,
			callee *ast.Expression,
			node *ast.Node,
		) {
			// A report requires at least one argument whose type could be unsafe;
			// skip the callee/signature type queries when none qualifies.
			if !slices.ContainsFunc(args, argumentCanBeUnsafe) {
				return
			}

			// ignore any-typed calls as these are caught by no-unsafe-call
			if utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(callee)) {
				return
			}

			signature := newFunctionSignature(ctx.TypeChecker, node)
			if signature == nil {
				panic("Expected to a signature resolved")
			}

			if ast.IsTaggedTemplateExpression(node) {
				// Consumes the first parameter (TemplateStringsArray) of the function called with TaggedTemplateExpression.
				signature.skipParameter()
			}

			for _, argument := range args {
				switch argument.Kind {
				// spreads consume
				case ast.KindSpreadElement:
					spreadArgType := ctx.TypeChecker.GetTypeAtLocation(argument.Expression())

					if utils.IsTypeAnyType(spreadArgType) {
						// foo(...any)
						ctx.ReportNode(argument, buildUnsafeSpreadMessage(describeType(spreadArgType)))
					} else if utils.IsTypeAnyArrayType(spreadArgType, ctx.TypeChecker) {
						// foo(...any[])

						// TODO - we could break down the spread and compare the array type against each argument
						ctx.ReportNode(argument, buildUnsafeArraySpreadMessage(describeTypeForSpread(spreadArgType)))
					} else if checker.IsTupleType(spreadArgType) {
						// foo(...[tuple1, tuple2])
						spreadTypeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, spreadArgType)
						for _, tupleType := range spreadTypeArguments {
							parameterType := signature.getNextParameterType()
							if parameterType == nil {
								continue
							}
							_, _, unsafe := utils.IsUnsafeAssignment(
								tupleType,
								parameterType,
								ctx.TypeChecker,
								// we can't pass the individual tuple members in here as this will most likely be a spread variable
								// not a spread array
								nil,
							)
							if unsafe {
								ctx.ReportNode(argument, buildUnsafeTupleSpreadMessage(describeTypeForTuple(tupleType), describeType(parameterType)))
							}
						}
						if checker.TupleType_combinedFlags(spreadArgType.Target().AsTupleType())&checker.ElementFlagsVariable != 0 {
							// the last element was a rest - so all remaining defined arguments can be considered "consumed"
							// all remaining arguments should be compared against the rest type (if one exists)
							signature.consumeRemainingArguments()
						}
					} else {
						// something that's iterable
						// handling this will be pretty complex - so we ignore it for now
						// TODO - handle generic iterable case
					}

				default:
					if !argumentCanBeUnsafe(argument) {
						signature.skipParameter()
						continue
					}

					parameterType := signature.getNextParameterType()
					if parameterType == nil {
						continue
					}

					argumentType := ctx.TypeChecker.GetTypeAtLocation(argument)
					_, _, unsafe := utils.IsUnsafeAssignment(
						argumentType,
						parameterType,
						ctx.TypeChecker,
						argument,
					)
					if unsafe {
						ctx.ReportNode(argument, buildUnsafeArgumentMessage(describeType(argumentType), describeType(parameterType)))
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				checkUnsafeArguments(node.Arguments(), node.Expression(), node)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				checkUnsafeArguments(node.Arguments(), node.Expression(), node)
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				expr := node.AsTaggedTemplateExpression()
				template := expr.Template
				if ast.IsTemplateExpression(template) {
					checkUnsafeArguments(utils.Map(template.AsTemplateExpression().TemplateSpans.Nodes, func(span *ast.Node) *ast.Node {
						return span.Expression()
					}), expr.Tag, node)
				}
			},
		}
	},
}
