package utils

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
)

type TypeOrValueSpecifierFrom uint8

const (
	TypeOrValueSpecifierFromFile TypeOrValueSpecifierFrom = iota
	TypeOrValueSpecifierFromLib
	TypeOrValueSpecifierFromPackage
)

type TypeOrValueSpecifier struct {
	From TypeOrValueSpecifierFrom
	Name []string
	// Can be used when From == TypeOrValueSpecifierFromFile
	Path string
	// Can be used when From == TypeOrValueSpecifierFromPackage
	Package string
}

// UnmarshalJSON implements json.Unmarshaler for TypeOrValueSpecifier.
// Handles both string specifiers and object specifiers with "from" field.
func (s *TypeOrValueSpecifier) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first (universal string specifier)
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = TypeOrValueSpecifier{
			From: TypeOrValueSpecifierFromFile, // Default to file for universal specifiers
			Name: []string{str},
		}
		return nil
	}

	// Otherwise, unmarshal as an object
	var raw struct {
		From    string      `json:"from"`
		Name    interface{} `json:"name"` // Can be string or []string
		Path    *string     `json:"path,omitempty"`
		Package *string     `json:"package,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("TypeOrValueSpecifier must be a string or object with 'from' field: %w", err)
	}

	// Parse the "from" field
	var from TypeOrValueSpecifierFrom
	switch raw.From {
	case "", "file":
		from = TypeOrValueSpecifierFromFile
	case "lib":
		from = TypeOrValueSpecifierFromLib
	case "package":
		from = TypeOrValueSpecifierFromPackage
	default:
		return fmt.Errorf("invalid 'from' field: %s (must be 'file', 'lib', or 'package')", raw.From)
	}

	// Parse the "name" field (can be string or []string)
	var names []string
	switch nameVal := raw.Name.(type) {
	case nil:
		// name field is optional for some specifier types
		names = []string{}
	case string:
		names = []string{nameVal}
	case []interface{}:
		names = make([]string, 0, len(nameVal))
		for _, n := range nameVal {
			if str, ok := n.(string); ok {
				names = append(names, str)
			} else {
				return errors.New("name array must contain only strings")
			}
		}
	default:
		return errors.New("name must be a string or array of strings")
	}

	// Build the result
	*s = TypeOrValueSpecifier{
		From: from,
		Name: names,
	}

	if raw.Path != nil {
		s.Path = *raw.Path
	}
	if raw.Package != nil {
		s.Package = *raw.Package
	}

	return nil
}

// MarshalJSON implements json.Marshaler for TypeOrValueSpecifier.
// Converts the struct back to JSON format compatible with UnmarshalJSON.
func (s TypeOrValueSpecifier) MarshalJSON() ([]byte, error) {
	// Convert From enum to string
	var fromStr string
	switch s.From {
	case TypeOrValueSpecifierFromFile:
		fromStr = "file"
	case TypeOrValueSpecifierFromLib:
		fromStr = "lib"
	case TypeOrValueSpecifierFromPackage:
		fromStr = "package"
	default:
		return nil, fmt.Errorf("invalid TypeOrValueSpecifierFrom value: %d", s.From)
	}

	// Build the output object
	output := map[string]interface{}{
		"from": fromStr,
		"name": s.Name,
	}

	if s.Path != "" {
		output["path"] = s.Path
	}
	if s.Package != "" {
		output["package"] = s.Package
	}

	return json.Marshal(output)
}

func typeMatchesStringSpecifier(
	t *checker.Type,
	names []string,
) bool {
	alias := checker.Type_alias(t)
	var symbol *ast.Symbol
	if alias == nil {
		symbol = checker.Type_symbol(t)
	} else {
		symbol = alias.Symbol()
	}

	if symbol != nil && slices.Contains(names, symbol.Name) {
		return true
	}

	if IsIntrinsicType(t) && slices.Contains(names, t.AsIntrinsicType().IntrinsicName()) {
		return true
	}

	return false
}

func typeDeclaredInFile(
	relativePath string,
	declarationFiles []*ast.SourceFile,
	program *compiler.Program,
) bool {
	cwd := program.Host().GetCurrentDirectory()
	if relativePath == "" {
		return Some(declarationFiles, func(f *ast.SourceFile) bool {
			return strings.HasPrefix(f.FileName(), cwd)
		})
	}
	absPath := tspath.GetNormalizedAbsolutePath(relativePath, cwd)
	return Some(declarationFiles, func(f *ast.SourceFile) bool {
		return f.FileName() == absPath
	})
}

func typeDeclaredInLib(
	declarationFiles []*ast.SourceFile,
	program *compiler.Program,
) bool {
	// Assertion: The type is not an error type.

	// Intrinsic type (i.e. string, number, boolean, etc) - Treat it as if it's from lib.
	if len(declarationFiles) == 0 {
		return true
	}
	return Some(declarationFiles, func(d *ast.SourceFile) bool {
		return IsSourceFileDefaultLibrary(program, d)
	})
}

func findParentModuleDeclaration(
	node *ast.Node,
) *ast.ModuleDeclaration {
	switch node.Kind {
	case ast.KindModuleDeclaration:
		decl := node.AsModuleDeclaration()
		if ast.IsStringLiteral(decl.Name()) {
			return decl
		}
		return nil
	case ast.KindSourceFile:
		return nil
	default:
		return findParentModuleDeclaration(node.Parent)
	}
}

func typeDeclaredInDeclareModule(
	packageName string,
	declarations []*ast.Node,
) bool {
	return Some(declarations, func(d *ast.Node) bool {
		parentModule := findParentModuleDeclaration(d)
		return parentModule != nil && parentModule.Name().Text() == packageName
	})
}

func typeDeclaredInDeclarationFile(
	packageName string,
	declarationFiles []*ast.SourceFile,
	program *compiler.Program,
) bool {
	// typesPackageName := ""
	//  // Handle scoped packages: if the name starts with @, remove it and replace / with __
	// slashIndex := strings.Index(packageName, "/")
	// if packageName[0] == '@' && slashIndex >= 0 {
	// 	typesPackageName = packageName[1:slashIndex] + "__" + packageName[slashIndex+1:]
	// }

	// TODO(port): there is no sourceFileToPackageName anymore
	// it looks like there is no other way to know sourceFile2PackageName,
	// other than set package name for ast.SourceFile in resolver

	return false

	// const matcher = new RegExp(`${packageName}|${typesPackageName}`);
	// return declarationFiles.some(declaration => {
	//   const packageIdName = program.sourceFileToPackageName.get(declaration.path);
	//   return (
	//     packageIdName != null &&
	//     matcher.test(packageIdName) &&
	//     program.isSourceFileFromExternalLibrary(declaration)
	//   );
	// });
}

func typeDeclaredInPackageDeclarationFile(
	packageName string,
	declarations []*ast.Node,
	declarationFiles []*ast.SourceFile,
	program *compiler.Program,
) bool {
	return typeDeclaredInDeclareModule(packageName, declarations) ||
		typeDeclaredInDeclarationFile(packageName, declarationFiles, program)
}

func typeMatchesSpecifier(
	t *checker.Type,
	specifier TypeOrValueSpecifier,
	program *compiler.Program,
) bool {
	if !typeMatchesStringSpecifier(t, specifier.Name) {
		return false
	}

	symbol := checker.Type_symbol(t)
	if symbol == nil {
		alias := checker.Type_alias(t)
		if alias != nil {
			symbol = alias.Symbol()
		}
	}
	var declarations []*ast.Node
	if symbol != nil {
		declarations = symbol.Declarations
	}
	declarationFiles := Map(declarations, func(d *ast.Node) *ast.SourceFile {
		return ast.GetSourceFileOfNode(d)
	})

	switch specifier.From {
	case TypeOrValueSpecifierFromFile:
		return typeDeclaredInFile(specifier.Path, declarationFiles, program)
	case TypeOrValueSpecifierFromLib:
		return typeDeclaredInLib(declarationFiles, program)
	case TypeOrValueSpecifierFromPackage:
		return typeDeclaredInPackageDeclarationFile(specifier.Package, declarations, declarationFiles, program)
	default:
		panic(fmt.Sprintf("unknown type specifier from: %v", specifier.From))
	}
}

// ConvertTypeOrValueSpecifier converts an interface{} (from JSON schema) to a TypeOrValueSpecifier struct.
// The input can be:
// - A string (universal string specifier - matches all names)
// - A map with "from" field indicating file/lib/package specifier
func ConvertTypeOrValueSpecifier(spec interface{}) (TypeOrValueSpecifier, bool) {
	// Handle string specifier
	if str, ok := spec.(string); ok {
		return TypeOrValueSpecifier{
			From: TypeOrValueSpecifierFromFile, // Default to file for universal specifiers
			Name: []string{str},
		}, true
	}

	// Handle object specifier
	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return TypeOrValueSpecifier{}, false
	}

	fromStr, ok := specMap["from"].(string)
	if !ok {
		return TypeOrValueSpecifier{}, false
	}

	var from TypeOrValueSpecifierFrom
	switch fromStr {
	case "file":
		from = TypeOrValueSpecifierFromFile
	case "lib":
		from = TypeOrValueSpecifierFromLib
	case "package":
		from = TypeOrValueSpecifierFromPackage
	default:
		return TypeOrValueSpecifier{}, false
	}

	// Extract name(s)
	var names []string
	switch nameVal := specMap["name"].(type) {
	case string:
		names = []string{nameVal}
	case []interface{}:
		names = make([]string, 0, len(nameVal))
		for _, n := range nameVal {
			if str, ok := n.(string); ok {
				names = append(names, str)
			}
		}
	default:
		return TypeOrValueSpecifier{}, false
	}

	result := TypeOrValueSpecifier{
		From: from,
		Name: names,
	}

	// Extract optional path (for file specifiers)
	if pathVal, ok := specMap["path"].(string); ok {
		result.Path = pathVal
	}

	// Extract optional package (for package specifiers)
	if pkgVal, ok := specMap["package"].(string); ok {
		result.Package = pkgVal
	}

	return result, true
}

func TypeMatchesSomeSpecifier(
	t *checker.Type,
	specifiers []TypeOrValueSpecifier,
	program *compiler.Program,
) bool {
	for _, typePart := range IntersectionTypeParts(t) {
		if IsIntrinsicErrorType(typePart) {
			continue
		}
		if Some(specifiers, func(s TypeOrValueSpecifier) bool {
			return typeMatchesSpecifier(t, s, program)
		}) {
			return true
		}
	}
	return false
}

func getStaticName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(node.AsPrivateIdentifier().Text, "#")
	case ast.KindStringLiteral:
		return node.Text()
	default:
		return ""
	}
}

func valueMatchesSpecifier(
	node *ast.Node,
	specifier TypeOrValueSpecifier,
	program *compiler.Program,
	t *checker.Type,
) bool {
	nodeName := getStaticName(node)
	if nodeName == "" {
		return false
	}

	// Check if the name matches
	if !slices.Contains(specifier.Name, nodeName) {
		return false
	}

	// Get the source file of the node
	sourceFile := ast.GetSourceFileOfNode(node)
	if sourceFile == nil {
		return false
	}

	switch specifier.From {
	case TypeOrValueSpecifierFromFile:
		// Check if declared in the specified file (or current file if path is empty)
		cwd := program.Host().GetCurrentDirectory()
		if specifier.Path == "" {
			// Empty path means current file (local to the file being linted)
			return strings.HasPrefix(sourceFile.FileName(), cwd)
		}
		absPath := tspath.GetNormalizedAbsolutePath(specifier.Path, cwd)
		return sourceFile.FileName() == absPath

	case TypeOrValueSpecifierFromLib:
		// Check if from a lib file
		return IsSourceFileDefaultLibrary(program, sourceFile)

	case TypeOrValueSpecifierFromPackage:
		// Check if from the specified package
		// For imports, we need to check the module specifier
		// Walk up to find import declaration
		current := node
		for current != nil {
			if current.Kind == ast.KindImportDeclaration {
				importDecl := current.AsImportDeclaration()
				if importDecl.ModuleSpecifier != nil {
					moduleSpec := importDecl.ModuleSpecifier.Text()
					// Strip quotes
					moduleSpec = strings.Trim(moduleSpec, "\"'")
					return moduleSpec == specifier.Package
				}
			}
			current = current.Parent
		}

		// Also check if the type's declarations are from a declare module
		if t != nil {
			symbol := checker.Type_symbol(t)
			if symbol != nil && len(symbol.Declarations) > 0 {
				return typeDeclaredInDeclareModule(specifier.Package, symbol.Declarations)
			}
		}

		return false

	default:
		panic(fmt.Sprintf("unknown value specifier from: %v", specifier.From))
	}
}

func ValueMatchesSomeSpecifier(
	node *ast.Node,
	specifiers []TypeOrValueSpecifier,
	program *compiler.Program,
	ty *checker.Type,
) bool {
	for _, s := range specifiers {
		if valueMatchesSpecifier(
			node,
			s,
			program,
			ty,
		) {
			return true
		}
	}
	return false
}
