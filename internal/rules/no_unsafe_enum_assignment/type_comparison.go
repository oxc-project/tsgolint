package no_unsafe_enum_assignment

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type typePair struct{ sender, receiver *checker.Type }

type comparisonVisit struct {
	typePair
	withinGenericMembers bool
}

const maximumTypeComparisonDepth = 5

type enumTypeChecker struct {
	checker      *checker.Checker
	genericTypes map[*checker.Type]bool
	mismatches   map[typePair]bool
}

func (c *enumTypeChecker) constraint(t *checker.Type) *checker.Type {
	if constraint := checker.Checker_getBaseConstraintOfType(c.checker, t); constraint != nil {
		return constraint
	}
	return t
}

func (c *enumTypeChecker) typeArguments(t *checker.Type) []*checker.Type {
	if utils.IsTypeFlagSet(t, checker.TypeFlagsObject) && checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0 {
		return checker.Checker_getTypeArguments(c.checker, t)
	}
	return nil
}

func (c *enumTypeChecker) isGeneric(t *checker.Type) bool {
	if generic, ok := c.genericTypes[t]; ok {
		return generic
	}
	c.genericTypes[t] = false
	generic := utils.IsTypeFlagSet(t, checker.TypeFlagsInstantiable)
	if !generic && utils.IsTypeFlagSet(t, checker.TypeFlagsUnionOrIntersection) {
		generic = slices.ContainsFunc(t.Types(), c.isGeneric)
	}
	if !generic && t.Alias() != nil {
		generic = slices.ContainsFunc(t.Alias().TypeArguments(), c.isGeneric)
	}
	if !generic {
		generic = slices.ContainsFunc(c.typeArguments(t), c.isGeneric)
	}
	c.genericTypes[t] = generic
	return generic
}

func (c *enumTypeChecker) mismatch(sender, receiver *checker.Type) bool {
	pair := typePair{sender, receiver}
	if mismatch, ok := c.mismatches[pair]; ok {
		return mismatch
	}
	mismatch := c.deepMismatch(sender, receiver, make(map[comparisonVisit]int), 0, false)
	c.mismatches[pair] = mismatch
	return mismatch
}

func (c *enumTypeChecker) deepMismatch(sender, receiver *checker.Type, visited map[comparisonVisit]int, depth int, withinGenericMembers bool) bool {
	if depth > maximumTypeComparisonDepth {
		return false
	}
	sender, receiver = c.constraint(sender), c.constraint(receiver)
	if sender == receiver {
		return false
	}
	visit := comparisonVisit{typePair{sender, receiver}, withinGenericMembers}
	if previousDepth, ok := visited[visit]; ok && previousDepth <= depth {
		return false
	}
	visited[visit] = depth
	if c.directMismatch(sender, receiver) {
		return true
	}
	senderArgs, receiverArgs := c.typeArguments(sender), c.typeArguments(receiver)
	if len(senderArgs) == len(receiverArgs) {
		for i, argument := range senderArgs {
			if c.deepMismatch(argument, receiverArgs[i], visited, depth+1, withinGenericMembers) {
				return true
			}
		}
	}
	generic := c.isGeneric(sender) || c.isGeneric(receiver)
	if generic && withinGenericMembers {
		return false
	}
	senderElement, receiverElement := c.checker.GetNumberIndexType(sender), c.checker.GetNumberIndexType(receiver)
	if senderElement != nil && receiverElement != nil && c.deepMismatch(senderElement, receiverElement, visited, depth+1, generic || withinGenericMembers) {
		return true
	}
	for _, receiverProperty := range checker.Checker_getPropertiesOfType(c.checker, receiver) {
		senderProperty := checker.Checker_getPropertyOfType(c.checker, sender, receiverProperty.Name)
		if senderProperty != nil && c.deepMismatch(checker.Checker_getTypeOfSymbol(c.checker, senderProperty), checker.Checker_getTypeOfSymbol(c.checker, receiverProperty), visited, depth+1, generic || withinGenericMembers) {
			return true
		}
	}
	return false
}

func primitiveLike(t *checker.Type, flags checker.TypeFlags) bool {
	return utils.Every(utils.UnionTypeParts(t), func(part *checker.Type) bool {
		return utils.Some(utils.IntersectionTypeParts(part), func(part *checker.Type) bool { return utils.IsTypeFlagSet(part, flags) })
	})
}

func (c *enumTypeChecker) directMismatch(sender, receiver *checker.Type) bool {
	receiverEnums := utils.GetEnumTypes(c.checker, receiver)
	var nonEnums []*checker.Type
	var numberEnum, stringEnum bool
	for _, part := range utils.UnionTypeParts(receiver) {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsEnumLike) {
			if utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLiteral) {
				numberEnum = true
			} else {
				stringEnum = true
			}
		}
		if len(utils.GetEnumTypes(c.checker, part)) == 0 {
			nonEnums = append(nonEnums, part)
		}
	}
	for _, part := range utils.UnionTypeParts(sender) {
		if !(numberEnum && primitiveLike(part, checker.TypeFlagsNumberLike) || stringEnum && primitiveLike(part, checker.TypeFlagsStringLike)) {
			continue
		}
		shared := false
		for _, senderEnum := range utils.GetEnumTypes(c.checker, part) {
			if slices.Contains(receiverEnums, senderEnum) {
				shared = true
				break
			}
		}
		if !shared && !slices.ContainsFunc(nonEnums, func(receiver *checker.Type) bool {
			return checker.Checker_isTypeAssignableTo(c.checker, part, receiver)
		}) {
			return true
		}
	}
	return false
}

func (c *enumTypeChecker) describe(types []*checker.Type) string {
	names := make(map[string]bool)
	visited := make(map[*checker.Type]int)
	var visit func(*checker.Type, int)
	visit = func(t *checker.Type, depth int) {
		if depth > maximumTypeComparisonDepth {
			return
		}
		t = c.constraint(t)
		if previousDepth, ok := visited[t]; ok && previousDepth <= depth {
			return
		}
		visited[t] = depth
		for _, enum := range utils.GetEnumTypes(c.checker, t) {
			names[c.checker.TypeToString(enum)] = true
		}
		for _, argument := range c.typeArguments(t) {
			visit(argument, depth+1)
		}
		if element := c.checker.GetNumberIndexType(t); element != nil {
			visit(element, depth+1)
		}
		for _, property := range checker.Checker_getPropertiesOfType(c.checker, t) {
			visit(checker.Checker_getTypeOfSymbol(c.checker, property), depth+1)
		}
	}
	for _, t := range types {
		visit(t, 0)
	}
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, "'"+name+"'")
	}
	slices.Sort(result)
	return strings.Join(result, ", ")
}

func mappedKeyConstraintTypes(typeChecker *checker.Checker, declaration *ast.Node) []*checker.Type {
	switch declaration.Kind {
	case ast.KindGetAccessor, ast.KindParameter, ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindVariableDeclaration:
	default:
		return nil
	}
	typeNode := declaration.Type()
	if typeNode == nil {
		return nil
	}
	if ast.IsTypeReferenceNode(typeNode) {
		t := checker.Checker_getTypeFromTypeNode(typeChecker, typeNode)
		if utils.IsObjectType(t) && checker.Type_objectFlags(t)&checker.ObjectFlagsMapped != 0 {
			if symbol := checker.Type_symbol(t); symbol != nil {
				for _, declaration := range symbol.Declarations {
					if ast.IsMappedTypeNode(declaration) && declaration.AsMappedTypeNode().NameType != nil {
						return nil
					}
				}
			}
			checker.Checker_getPropertiesOfType(typeChecker, t)
			constraint := checker.MappedType_constraintType(t.AsMappedType())
			if constraint != nil {
				constrained := checker.Checker_getBaseConstraintOfType(typeChecker, constraint)
				if constrained == nil {
					constrained = constraint
				}
				if len(utils.GetEnumTypes(typeChecker, constrained)) > 0 {
					return []*checker.Type{constraint}
				}
			}
		}
	}
	if ast.IsMappedTypeNode(typeNode) {
		constraint := typeNode.AsMappedTypeNode().TypeParameter.AsTypeParameterDeclaration().Constraint
		if constraint != nil {
			return []*checker.Type{checker.Checker_getTypeFromTypeNode(typeChecker, constraint)}
		}
	}
	var result []*checker.Type
	if ast.IsTypeLiteralNode(typeNode) {
		for _, member := range typeNode.AsTypeLiteralNode().Members.Nodes {
			if name := member.Name(); name != nil && ast.IsComputedPropertyName(name) {
				result = append(result, typeChecker.GetTypeAtLocation(name.Expression()))
			}
		}
	}
	return result
}
