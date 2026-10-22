package no_unsafe_enum_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type functionSignature struct {
	parameters []*checker.Type
	rest       []*checker.Type
	restIndex  int
	index      int
	consumed   bool
}

func newFunctionSignature(c *enumTypeChecker, node *ast.Node) *functionSignature {
	signature := checker.Checker_getResolvedSignature(c.checker, node, nil, checker.CheckModeNormal)
	if signature == nil {
		return nil
	}
	result := &functionSignature{}
	for i, parameter := range checker.Signature_parameters(signature) {
		t := c.checker.GetTypeOfSymbolAtLocation(parameter, node)
		constrained := c.constraint(t)
		if len(parameter.Declarations) > 0 && utils.IsRestParameterDeclaration(parameter.Declarations[0]) {
			result.restIndex = i
			if checker.IsTupleType(constrained) {
				result.rest = c.typeArguments(constrained)
			} else if element := c.checker.GetNumberIndexType(constrained); element != nil {
				result.rest = []*checker.Type{element}
			} else {
				result.rest = []*checker.Type{constrained}
			}
			break
		}
		result.parameters = append(result.parameters, t)
	}
	return result
}

func (s *functionSignature) next() *checker.Type {
	index := s.index
	s.index++
	if index < len(s.parameters) && !s.consumed {
		return s.parameters[index]
	}
	if len(s.rest) == 0 {
		return nil
	}
	if s.consumed || index-s.restIndex >= len(s.rest) {
		return s.rest[len(s.rest)-1]
	}
	return s.rest[index-s.restIndex]
}
