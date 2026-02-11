package naming_convention

import (
	"context"
	"fmt"
	"math/bits"
	"regexp"
	"slices"
	"strings"
	"unicode"
	_ "unsafe"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

//go:linkname checkerCheckSourceFile github.com/microsoft/typescript-go/internal/checker.(*Checker).checkSourceFile
func checkerCheckSourceFile(recv *checker.Checker, ctx context.Context, sourceFile *ast.SourceFile)

// Selector bit flags for matching AST node types.
type Selector int

const (
	SelectorVariable              Selector = 1 << 0  // 1
	SelectorFunction              Selector = 1 << 1  // 2
	SelectorParameter             Selector = 1 << 2  // 4
	SelectorParameterProperty     Selector = 1 << 3  // 8
	SelectorClassicAccessor       Selector = 1 << 4  // 16
	SelectorEnumMember            Selector = 1 << 5  // 32
	SelectorClassMethod           Selector = 1 << 6  // 64
	SelectorObjectLiteralMethod   Selector = 1 << 7  // 128
	SelectorTypeMethod            Selector = 1 << 8  // 256
	SelectorClassProperty         Selector = 1 << 9  // 512
	SelectorObjectLiteralProperty Selector = 1 << 10 // 1024
	SelectorTypeProperty          Selector = 1 << 11 // 2048
	SelectorAutoAccessor          Selector = 1 << 12 // 4096
	SelectorClass                 Selector = 1 << 13 // 8192
	SelectorInterface             Selector = 1 << 14 // 16384
	SelectorTypeAlias             Selector = 1 << 15 // 32768
	SelectorEnum                  Selector = 1 << 16 // 65536
	SelectorTypeParameter         Selector = 1 << 17 // 131072
	SelectorImport                Selector = 1 << 18 // 262144
)

// Meta selectors (composites).
const (
	SelectorDefault      Selector = -1
	SelectorVariableLike Selector = SelectorVariable | SelectorFunction | SelectorParameter
	SelectorMemberLike   Selector = SelectorParameterProperty | SelectorClassicAccessor | SelectorEnumMember |
		SelectorClassMethod | SelectorObjectLiteralMethod | SelectorTypeMethod |
		SelectorClassProperty | SelectorObjectLiteralProperty | SelectorTypeProperty | SelectorAutoAccessor
	SelectorTypeLike Selector = SelectorClass | SelectorInterface | SelectorTypeAlias | SelectorEnum | SelectorTypeParameter
	SelectorMethod   Selector = SelectorClassMethod | SelectorObjectLiteralMethod | SelectorTypeMethod
	SelectorProperty Selector = SelectorClassProperty | SelectorObjectLiteralProperty | SelectorTypeProperty
	SelectorAccessor Selector = SelectorClassicAccessor | SelectorAutoAccessor
)

// Modifier bit flags for identifier modifiers.
type Modifier int

const (
	ModifierConst          Modifier = 1 << 0  // 1
	ModifierReadonly       Modifier = 1 << 1  // 2
	ModifierStatic         Modifier = 1 << 2  // 4
	ModifierPublic         Modifier = 1 << 3  // 8
	ModifierProtected      Modifier = 1 << 4  // 16
	ModifierPrivate        Modifier = 1 << 5  // 32
	ModifierHashPrivate    Modifier = 1 << 6  // 64
	ModifierAbstract       Modifier = 1 << 7  // 128
	ModifierDestructured   Modifier = 1 << 8  // 256
	ModifierGlobal         Modifier = 1 << 9  // 512
	ModifierExported       Modifier = 1 << 10 // 1024
	ModifierUnused         Modifier = 1 << 11 // 2048
	ModifierRequiresQuotes Modifier = 1 << 12 // 4096
	ModifierOverride       Modifier = 1 << 13 // 8192
	ModifierAsync          Modifier = 1 << 14 // 16384
	ModifierDefaultImport  Modifier = 1 << 15 // 32768
	ModifierNamespace      Modifier = 1 << 16 // 65536
)

// TypeModifier bit flags for type-based matching.
type TypeModifier int

const (
	TypeModifierBoolean  TypeModifier = 1 << 17 // 131072
	TypeModifierString   TypeModifier = 1 << 18 // 262144
	TypeModifierNumber   TypeModifier = 1 << 19 // 524288
	TypeModifierFunction TypeModifier = 1 << 20 // 1048576
	TypeModifierArray    TypeModifier = 1 << 21 // 2097152
)

// PredefinedFormat represents naming format options.
type PredefinedFormat int

const (
	PredefinedFormatCamelCase        PredefinedFormat = 1
	PredefinedFormatStrictCamelCase  PredefinedFormat = 2
	PredefinedFormatPascalCase       PredefinedFormat = 3
	PredefinedFormatStrictPascalCase PredefinedFormat = 4
	PredefinedFormatSnakeCase        PredefinedFormat = 5
	PredefinedFormatUpperCase        PredefinedFormat = 6
)

// UnderscoreOption for leading/trailing underscore handling.
type UnderscoreOption int

const (
	UnderscoreForbid              UnderscoreOption = 1
	UnderscoreAllow               UnderscoreOption = 2
	UnderscoreRequire             UnderscoreOption = 3
	UnderscoreRequireDouble       UnderscoreOption = 4
	UnderscoreAllowDouble         UnderscoreOption = 5
	UnderscoreAllowSingleOrDouble UnderscoreOption = 6
)

// normalizedSelector is a processed selector ready for matching.
type normalizedSelector struct {
	selector           Selector
	modifiers          Modifier
	types              TypeModifier
	filter             *normalizedFilter
	formats            []PredefinedFormat
	custom             *normalizedMatchRegex
	leadingUnderscore  UnderscoreOption
	trailingUnderscore UnderscoreOption
	prefix             []string
	suffix             []string
	weight             int
	selectorName       string
	hasFormats         bool // true when format was explicitly set (even if empty slice); false means format was null
}

// selectorGroups provides O(1) lookup by individual selector, indexed by bit position.
type selectorGroups [][]normalizedSelector

func selectorIndex(sel Selector) int {
	return bits.TrailingZeros(uint(sel))
}

// normalizedFilter is a compiled filter regex.
type normalizedFilter struct {
	regex *regexp.Regexp
	match bool
}

// normalizedMatchRegex is a compiled custom regex.
type normalizedMatchRegex struct {
	regex *regexp.Regexp
	match bool
}

// Message builder functions

func buildUnexpectedUnderscoreMessage(selectorName, name, position string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedUnderscore",
		Description: fmt.Sprintf("%s name `%s` must not have a %s underscore.", selectorName, name, position),
	}
}

func buildMissingUnderscoreMessage(selectorName, name, position string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingUnderscore",
		Description: fmt.Sprintf("%s name `%s` must have a %s underscore.", selectorName, name, position),
	}
}

func buildMissingDoubleUnderscoreMessage(selectorName, name, position string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingUnderscore",
		Description: fmt.Sprintf("%s name `%s` must have a %s double underscore.", selectorName, name, position),
	}
}

func buildMissingAffixMessage(selectorName, name, affixType string, affixes []string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAffix",
		Description: fmt.Sprintf("%s name `%s` must have one of the following %ses: %s", selectorName, name, affixType, formatAffixList(affixes)),
	}
}

func buildSatisfyCustomMessage(selectorName, name, matchStr, regex string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "satisfyCustom",
		Description: fmt.Sprintf("%s name `%s` must %s the RegExp `%s`.", selectorName, name, matchStr, regex),
	}
}

func buildDoesNotMatchFormatMessage(selectorName, name, formatNames string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "doesNotMatchFormat",
		Description: fmt.Sprintf("%s name `%s` must match one of the following formats: %s", selectorName, name, formatNames),
	}
}

func buildDoesNotMatchFormatTrimmedMessage(selectorName, originalName, processedName, formatNames string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "doesNotMatchFormatTrimmed",
		Description: fmt.Sprintf("%s name `%s` trimmed as `%s` must match one of the following formats: %s", selectorName, originalName, processedName, formatNames),
	}
}

var selectorNameMap = map[string]Selector{
	"default":               SelectorDefault,
	"variable":              SelectorVariable,
	"function":              SelectorFunction,
	"parameter":             SelectorParameter,
	"parameterProperty":     SelectorParameterProperty,
	"classicAccessor":       SelectorClassicAccessor,
	"enumMember":            SelectorEnumMember,
	"classMethod":           SelectorClassMethod,
	"objectLiteralMethod":   SelectorObjectLiteralMethod,
	"typeMethod":            SelectorTypeMethod,
	"classProperty":         SelectorClassProperty,
	"objectLiteralProperty": SelectorObjectLiteralProperty,
	"typeProperty":          SelectorTypeProperty,
	"autoAccessor":          SelectorAutoAccessor,
	"class":                 SelectorClass,
	"interface":             SelectorInterface,
	"typeAlias":             SelectorTypeAlias,
	"enum":                  SelectorEnum,
	"typeParameter":         SelectorTypeParameter,
	"import":                SelectorImport,
	// Meta selectors
	"variableLike": SelectorVariableLike,
	"memberLike":   SelectorMemberLike,
	"typeLike":     SelectorTypeLike,
	"method":       SelectorMethod,
	"property":     SelectorProperty,
	"accessor":     SelectorAccessor,
}

var modifierNameMap = map[string]Modifier{
	"const":          ModifierConst,
	"readonly":       ModifierReadonly,
	"static":         ModifierStatic,
	"public":         ModifierPublic,
	"protected":      ModifierProtected,
	"private":        ModifierPrivate,
	"#private":       ModifierHashPrivate,
	"abstract":       ModifierAbstract,
	"destructured":   ModifierDestructured,
	"global":         ModifierGlobal,
	"exported":       ModifierExported,
	"unused":         ModifierUnused,
	"requiresQuotes": ModifierRequiresQuotes,
	"override":       ModifierOverride,
	"async":          ModifierAsync,
	"default":        ModifierDefaultImport,
	"namespace":      ModifierNamespace,
}

var typeModifierNameMap = map[string]TypeModifier{
	"boolean":  TypeModifierBoolean,
	"string":   TypeModifierString,
	"number":   TypeModifierNumber,
	"function": TypeModifierFunction,
	"array":    TypeModifierArray,
}

var formatNameMap = map[string]PredefinedFormat{
	"camelCase":        PredefinedFormatCamelCase,
	"strictCamelCase":  PredefinedFormatStrictCamelCase,
	"PascalCase":       PredefinedFormatPascalCase,
	"StrictPascalCase": PredefinedFormatStrictPascalCase,
	"snake_case":       PredefinedFormatSnakeCase,
	"UPPER_CASE":       PredefinedFormatUpperCase,
}

var underscoreNameMap = map[string]UnderscoreOption{
	"forbid":              UnderscoreForbid,
	"allow":               UnderscoreAllow,
	"require":             UnderscoreRequire,
	"requireDouble":       UnderscoreRequireDouble,
	"allowDouble":         UnderscoreAllowDouble,
	"allowSingleOrDouble": UnderscoreAllowSingleOrDouble,
}

var formatNameForDisplay = map[PredefinedFormat]string{
	PredefinedFormatCamelCase:        "camelCase",
	PredefinedFormatStrictCamelCase:  "strictCamelCase",
	PredefinedFormatPascalCase:       "PascalCase",
	PredefinedFormatStrictPascalCase: "StrictPascalCase",
	PredefinedFormatSnakeCase:        "snake_case",
	PredefinedFormatUpperCase:        "UPPER_CASE",
}

var selectorTypeString = map[Selector]string{
	SelectorVariable:              "variable",
	SelectorFunction:              "function",
	SelectorParameter:             "parameter",
	SelectorParameterProperty:     "parameterProperty",
	SelectorClassicAccessor:       "classicAccessor",
	SelectorEnumMember:            "enumMember",
	SelectorClassMethod:           "classMethod",
	SelectorObjectLiteralMethod:   "objectLiteralMethod",
	SelectorTypeMethod:            "typeMethod",
	SelectorClassProperty:         "classProperty",
	SelectorObjectLiteralProperty: "objectLiteralProperty",
	SelectorTypeProperty:          "typeProperty",
	SelectorAutoAccessor:          "autoAccessor",
	SelectorClass:                 "class",
	SelectorInterface:             "interface",
	SelectorTypeAlias:             "typeAlias",
	SelectorEnum:                  "enum",
	SelectorTypeParameter:         "typeParameter",
	SelectorImport:                "import",
}

var individualSelectors = []Selector{
	SelectorVariable,
	SelectorFunction,
	SelectorParameter,
	SelectorParameterProperty,
	SelectorClassicAccessor,
	SelectorEnumMember,
	SelectorClassMethod,
	SelectorObjectLiteralMethod,
	SelectorTypeMethod,
	SelectorClassProperty,
	SelectorObjectLiteralProperty,
	SelectorTypeProperty,
	SelectorAutoAccessor,
	SelectorClass,
	SelectorInterface,
	SelectorTypeAlias,
	SelectorEnum,
	SelectorTypeParameter,
	SelectorImport,
}

// Format checking functions

func checkFormat(format PredefinedFormat, name string) bool {
	switch format {
	case PredefinedFormatCamelCase:
		return isCamelCase(name)
	case PredefinedFormatStrictCamelCase:
		return isStrictCamelCase(name)
	case PredefinedFormatPascalCase:
		return isPascalCase(name)
	case PredefinedFormatStrictPascalCase:
		return isStrictPascalCase(name)
	case PredefinedFormatSnakeCase:
		return isSnakeCase(name)
	case PredefinedFormatUpperCase:
		return isUpperCase(name)
	}
	return false
}

func isCamelCase(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 && !unicode.IsLower(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isStrictCamelCase(name string) bool {
	return isCamelCase(name) && !hasConsecutiveUppercase(name)
}

func isPascalCase(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 && !unicode.IsUpper(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isStrictPascalCase(name string) bool {
	return isPascalCase(name) && !hasConsecutiveUppercase(name)
}

func isSnakeCase(name string) bool {
	if name == "" {
		return false
	}
	if name[0] == '_' || name[len(name)-1] == '_' {
		return false
	}
	prevUnderscore := false
	for _, r := range name {
		if r == '_' {
			if prevUnderscore {
				return false
			}
			prevUnderscore = true
			continue
		}
		prevUnderscore = false
		if !unicode.IsLower(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isUpperCase(name string) bool {
	if name == "" {
		return false
	}
	if name[0] == '_' || name[len(name)-1] == '_' {
		return false
	}
	prevUnderscore := false
	for _, r := range name {
		if r == '_' {
			if prevUnderscore {
				return false
			}
			prevUnderscore = true
			continue
		}
		prevUnderscore = false
		if !unicode.IsUpper(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func hasConsecutiveUppercase(name string) bool {
	prevUpper := false
	for _, r := range name {
		if unicode.IsUpper(r) {
			if prevUpper {
				return true
			}
			prevUpper = true
		} else {
			prevUpper = false
		}
	}
	return false
}

// Option normalization

func normalizeOptions(rawOptions []NamingConventionOption) selectorGroups {
	var all []normalizedSelector
	for _, opt := range rawOptions {
		selectors := parseSelectorNames(opt.Selector)
		modifiers := parseModifiers(opt.Modifiers)
		types := parseTypeModifiers(opt.Types)
		formats, hasFormats := parseFormats(opt.Format)
		leadingUnderscore := parseUnderscoreOption(opt.LeadingUnderscore)
		trailingUnderscore := parseUnderscoreOption(opt.TrailingUnderscore)
		filter := parseFilter(opt.Filter)
		custom := parseCustomRegex(opt.Custom)

		for _, sel := range selectors {
			expandedSelectors := expandSelector(sel)
			for _, expandedSel := range expandedSelectors {
				weight := calculateWeight(expandedSel, sel, modifiers, types, filter)
				all = append(all, normalizedSelector{
					selector:           expandedSel,
					modifiers:          modifiers,
					types:              types,
					filter:             filter,
					formats:            formats,
					custom:             custom,
					leadingUnderscore:  leadingUnderscore,
					trailingUnderscore: trailingUnderscore,
					prefix:             opt.Prefix,
					suffix:             opt.Suffix,
					weight:             weight,
					selectorName:       selectorTypeString[expandedSel],
					hasFormats:         hasFormats,
				})
			}
		}
	}

	// Sort by weight descending (most specific first)
	slices.SortStableFunc(all, func(a, b normalizedSelector) int {
		return b.weight - a.weight
	})

	// Group by selector for O(1) lookup in validateName
	groups := make(selectorGroups, len(individualSelectors))
	for i := range all {
		idx := selectorIndex(all[i].selector)
		groups[idx] = append(groups[idx], all[i])
	}

	return groups
}

func hasTypeModifierSelectors(groups selectorGroups) bool {
	for i := range groups {
		for j := range groups[i] {
			if groups[i][j].types != 0 {
				return true
			}
		}
	}
	return false
}

func hasModifierInSelectors(groups selectorGroups, mod Modifier) bool {
	for i := range groups {
		for j := range groups[i] {
			if groups[i][j].modifiers&mod != 0 {
				return true
			}
		}
	}
	return false
}

func parseSelectorNames(selectorRaw any) []Selector {
	switch v := selectorRaw.(type) {
	case string:
		if sel, ok := selectorNameMap[v]; ok {
			return []Selector{sel}
		}
	case []any:
		var selectors []Selector
		for _, item := range v {
			if s, ok := item.(string); ok {
				if sel, ok := selectorNameMap[s]; ok {
					selectors = append(selectors, sel)
				}
			}
		}
		return selectors
	case []string:
		var selectors []Selector
		for _, s := range v {
			if sel, ok := selectorNameMap[s]; ok {
				selectors = append(selectors, sel)
			}
		}
		return selectors
	}
	return []Selector{SelectorDefault}
}

func parseModifiers(modifiers []string) Modifier {
	var result Modifier
	for _, m := range modifiers {
		if mod, ok := modifierNameMap[m]; ok {
			result |= mod
		}
	}
	return result
}

func parseTypeModifiers(types []string) TypeModifier {
	var result TypeModifier
	for _, t := range types {
		if tm, ok := typeModifierNameMap[t]; ok {
			result |= tm
		}
	}
	return result
}

func parseFormats(formats *[]string) ([]PredefinedFormat, bool) {
	if formats == nil {
		return nil, false
	}
	var result []PredefinedFormat
	for _, f := range *formats {
		if pf, ok := formatNameMap[f]; ok {
			result = append(result, pf)
		}
	}
	return result, true
}

func parseUnderscoreOption(opt *string) UnderscoreOption {
	if opt == nil {
		return 0
	}
	if uo, ok := underscoreNameMap[*opt]; ok {
		return uo
	}
	return 0
}

func parseFilter(filterRaw any) *normalizedFilter {
	if filterRaw == nil {
		return nil
	}
	switch v := filterRaw.(type) {
	case string:
		compiled, err := regexp.Compile(v)
		if err != nil {
			return nil
		}
		return &normalizedFilter{regex: compiled, match: true}
	case map[string]any:
		regexStr, _ := v["regex"].(string)
		match, ok := v["match"].(bool)
		if !ok {
			match = true
		}
		compiled, err := regexp.Compile(regexStr)
		if err != nil {
			return nil
		}
		return &normalizedFilter{regex: compiled, match: match}
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		var mr MatchRegex
		if err := json.Unmarshal(data, &mr); err != nil {
			var s string
			if err := json.Unmarshal(data, &s); err != nil {
				return nil
			}
			compiled, err := regexp.Compile(s)
			if err != nil {
				return nil
			}
			return &normalizedFilter{regex: compiled, match: true}
		}
		compiled, err := regexp.Compile(mr.Regex)
		if err != nil {
			return nil
		}
		return &normalizedFilter{regex: compiled, match: mr.Match}
	}
}

func parseCustomRegex(custom *MatchRegex) *normalizedMatchRegex {
	if custom == nil {
		return nil
	}
	compiled, err := regexp.Compile(custom.Regex)
	if err != nil {
		return nil
	}
	return &normalizedMatchRegex{regex: compiled, match: custom.Match}
}

func expandSelector(sel Selector) []Selector {
	if sel == SelectorDefault {
		return individualSelectors
	}
	if isMetaSelector(sel) {
		result := make([]Selector, 0, bits.OnesCount(uint(sel)))
		for _, individual := range individualSelectors {
			if sel&individual != 0 {
				result = append(result, individual)
			}
		}
		return result
	}
	return []Selector{sel}
}

func isMetaSelector(sel Selector) bool {
	if sel == SelectorDefault {
		return true
	}
	return bits.OnesCount(uint(sel)) > 1
}

// isGroupMetaSelector returns true for the top-level meta selectors
// (memberLike, variableLike, typeLike) which are less specific than
// sub-group meta selectors (method, property, accessor).
func isGroupMetaSelector(sel Selector) bool {
	return sel == SelectorMemberLike || sel == SelectorVariableLike || sel == SelectorTypeLike
}

// calculateWeight determines the specificity of a selector configuration.
// Higher weight = more specific = should be checked first.
// Specificity tiers: default(0) < group meta(1) < sub-group meta(2) < individual(3) < modifiers < type modifiers < filter
func calculateWeight(expandedSel Selector, originalSel Selector, modifiers Modifier, types TypeModifier, filter *normalizedFilter) int {
	weight := 0

	// Four tiers of selector specificity at bits 0-1:
	// - default = 0
	// - group meta selectors (memberLike, variableLike, typeLike) = 1
	// - sub-group meta selectors (method, property, accessor) = 2
	// - individual selectors (classMethod, objectLiteralMethod, etc.) = 3
	if originalSel == SelectorDefault {
		// weight stays 0
	} else if isGroupMetaSelector(originalSel) {
		weight |= 1
	} else if isMetaSelector(originalSel) {
		weight |= 2
	} else {
		weight |= 3
	}

	// Each modifier adds weight (bits 2+)
	weight |= bits.OnesCount(uint(modifiers)) << 2

	// Type modifiers add more weight (bits 9+)
	weight |= bits.OnesCount(uint(types)) << 9

	// Filter adds the most weight
	if filter != nil {
		weight |= 1 << 30
	}

	return weight
}

// Validation pipeline

// selectorsWithTypesMask is the set of selectors that support type annotations.
const selectorsWithTypesMask Selector = SelectorVariable | SelectorParameter | SelectorParameterProperty |
	SelectorClassProperty | SelectorObjectLiteralProperty | SelectorTypeProperty |
	SelectorClassicAccessor | SelectorAutoAccessor

func validateName(
	name string,
	groups selectorGroups,
	nodeSelector Selector,
	nodeModifiers Modifier,
	nodeTypes TypeModifier,
) *rule.RuleMessage {
	selectors := groups[selectorIndex(nodeSelector)]
	if len(selectors) == 0 {
		return nil
	}
	for i := range selectors {
		sel := &selectors[i]

		// Check filter
		if sel.filter != nil {
			if sel.filter.regex.MatchString(name) != sel.filter.match {
				continue
			}
		}

		// Check modifiers
		if sel.modifiers != 0 && nodeModifiers&sel.modifiers != sel.modifiers {
			continue
		}

		// Check types — only for selectors that support type annotations
		if sel.types != 0 && nodeSelector&selectorsWithTypesMask != 0 && nodeTypes&sel.types == 0 {
			continue
		}

		// Found matching selector - run format validation
		return runFormatValidation(name, sel)
	}

	return nil
}

func runFormatValidation(name string, sel *normalizedSelector) *rule.RuleMessage {
	processedName := name
	selectorName := sel.selectorName

	if sel.leadingUnderscore != 0 {
		stripped, msg := validateUnderscore("leading", processedName, sel.leadingUnderscore, selectorName)
		if msg != nil {
			return msg
		}
		processedName = stripped
	}

	if sel.trailingUnderscore != 0 {
		stripped, msg := validateUnderscore("trailing", processedName, sel.trailingUnderscore, selectorName)
		if msg != nil {
			return msg
		}
		processedName = stripped
	}

	if len(sel.prefix) > 0 {
		stripped, msg := validateAffix("prefix", processedName, sel.prefix, selectorName)
		if msg != nil {
			return msg
		}
		processedName = stripped
	}

	if len(sel.suffix) > 0 {
		stripped, msg := validateAffix("suffix", processedName, sel.suffix, selectorName)
		if msg != nil {
			return msg
		}
		processedName = stripped
	}

	if sel.custom != nil {
		if msg := validateCustomRegex(processedName, sel.custom, selectorName); msg != nil {
			return msg
		}
	}

	if sel.hasFormats && len(sel.formats) > 0 {
		if msg := validatePredefinedFormat(name, processedName, sel.formats, selectorName); msg != nil {
			return msg
		}
	}

	return nil
}

func validateUnderscore(position string, name string, option UnderscoreOption, selectorName string) (string, *rule.RuleMessage) {
	var count int
	if position == "leading" {
		count = countLeadingUnderscores(name)
	} else {
		count = countTrailingUnderscores(name)
	}

	switch option {
	case UnderscoreForbid:
		if count > 0 {
			msg := buildUnexpectedUnderscoreMessage(selectorName, name, position)
			return "", &msg
		}
		return name, nil

	case UnderscoreAllow:
		if position == "leading" {
			if count > 0 {
				return name[1:], nil
			}
		} else {
			if count > 0 {
				return name[:len(name)-1], nil
			}
		}
		return name, nil

	case UnderscoreRequire:
		if count == 0 {
			msg := buildMissingUnderscoreMessage(selectorName, name, position)
			return "", &msg
		}
		if position == "leading" {
			return name[1:], nil
		}
		return name[:len(name)-1], nil

	case UnderscoreRequireDouble:
		if count < 2 {
			msg := buildMissingDoubleUnderscoreMessage(selectorName, name, position)
			return "", &msg
		}
		if position == "leading" {
			return name[2:], nil
		}
		return name[:len(name)-2], nil

	case UnderscoreAllowDouble:
		if count == 2 {
			if position == "leading" {
				return name[2:], nil
			}
			return name[:len(name)-2], nil
		}
		if count > 0 {
			msg := buildUnexpectedUnderscoreMessage(selectorName, name, position)
			return "", &msg
		}
		return name, nil

	case UnderscoreAllowSingleOrDouble:
		if count == 1 {
			if position == "leading" {
				return name[1:], nil
			}
			return name[:len(name)-1], nil
		}
		if count == 2 {
			if position == "leading" {
				return name[2:], nil
			}
			return name[:len(name)-2], nil
		}
		if count > 2 {
			msg := buildUnexpectedUnderscoreMessage(selectorName, name, position)
			return "", &msg
		}
		return name, nil
	}

	return name, nil
}

func validateAffix(affixType string, name string, affixes []string, selectorName string) (string, *rule.RuleMessage) {
	for _, affix := range affixes {
		if affixType == "prefix" {
			if strings.HasPrefix(name, affix) {
				return name[len(affix):], nil
			}
		} else {
			if strings.HasSuffix(name, affix) {
				return name[:len(name)-len(affix)], nil
			}
		}
	}
	msg := buildMissingAffixMessage(selectorName, name, affixType, affixes)
	return "", &msg
}

func validateCustomRegex(name string, custom *normalizedMatchRegex, selectorName string) *rule.RuleMessage {
	if custom.regex.MatchString(name) != custom.match {
		matchStr := "match"
		if !custom.match {
			matchStr = "not match"
		}
		msg := buildSatisfyCustomMessage(selectorName, name, matchStr, custom.regex.String())
		return &msg
	}
	return nil
}

func validatePredefinedFormat(originalName string, processedName string, formats []PredefinedFormat, selectorName string) *rule.RuleMessage {
	for _, format := range formats {
		if checkFormat(format, processedName) {
			return nil
		}
	}

	formatNames := make([]string, len(formats))
	for i, f := range formats {
		formatNames[i] = formatNameForDisplay[f]
	}
	joined := strings.Join(formatNames, ", ")

	if originalName == processedName {
		msg := buildDoesNotMatchFormatMessage(selectorName, originalName, joined)
		return &msg
	}
	msg := buildDoesNotMatchFormatTrimmedMessage(selectorName, originalName, processedName, joined)
	return &msg
}

func countLeadingUnderscores(name string) int {
	i := 0
	for i < len(name) && name[i] == '_' {
		i++
	}
	return i
}

func countTrailingUnderscores(name string) int {
	count := 0
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '_' {
			count++
		} else {
			break
		}
	}
	return count
}

func formatAffixList(affixes []string) string {
	quoted := make([]string, len(affixes))
	for i, a := range affixes {
		quoted[i] = "`" + a + "`"
	}
	return strings.Join(quoted, ", ")
}

// Default options and rule definition

var defaultOptions = []NamingConventionOption{
	{
		Selector:           "default",
		Format:             &[]string{"camelCase"},
		LeadingUnderscore:  strPtr("allow"),
		TrailingUnderscore: strPtr("allow"),
	},
	{
		Selector: "import",
		Format:   &[]string{"camelCase", "PascalCase"},
	},
	{
		Selector:           "variable",
		Format:             &[]string{"camelCase", "UPPER_CASE"},
		LeadingUnderscore:  strPtr("allow"),
		TrailingUnderscore: strPtr("allow"),
	},
	{
		Selector: "typeLike",
		Format:   &[]string{"PascalCase"},
	},
}

func strPtr(s string) *string { return &s }

var NamingConventionRule = rule.Rule{
	Name: "naming-convention",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		selectorOptions := parseOptions(options)
		groups := normalizeOptions(selectorOptions)
		needsTypeInfo := ctx.TypeChecker != nil && hasTypeModifierSelectors(groups)
		needsUnusedCheck := ctx.TypeChecker != nil && hasModifierInSelectors(groups, ModifierUnused)

		// Force full type checking on this checker so that isReferenced returns correct results
		if needsUnusedCheck {
			checkerCheckSourceFile(ctx.TypeChecker, context.Background(), ctx.SourceFile)
		}

		// Build a map of names exported via `export { ... }` blocks
		exportedViaBlock := buildExportedViaBlockMap(ctx.SourceFile)

		report := func(node *ast.Node, name string, nodeSelector Selector, nodeModifiers Modifier) {
			var nodeTypes TypeModifier
			if needsTypeInfo {
				nodeTypes = detectTypeModifiers(ctx, node)
			}

			if msg := validateName(name, groups, nodeSelector, nodeModifiers, nodeTypes); msg != nil {
				ctx.ReportNode(node, *msg)
			}
		}

		unusedModifier := func(nameNode *ast.Node) Modifier {
			if needsUnusedCheck {
				return detectUnusedModifier(ctx, nameNode, exportedViaBlock)
			}
			return 0
		}

		processNode := func(node *ast.Node, nodeSelector Selector, baseModifiers Modifier) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}

			if ast.IsComputedPropertyName(nameNode) {
				return
			}

			var name string
			isStringLiteralName := false
			if ast.IsIdentifier(nameNode) {
				name = nameNode.AsIdentifier().Text
			} else if ast.IsPrivateIdentifier(nameNode) {
				name = nameNode.Text()
			} else if ast.IsStringLiteral(nameNode) {
				name = nameNode.AsStringLiteral().Text
				baseModifiers |= ModifierRequiresQuotes
				isStringLiteralName = true
			} else {
				name = nameNode.Text()
			}

			if name == "" && !isStringLiteralName {
				return
			}

			report(nameNode, name, nodeSelector, baseModifiers)
		}

		listeners := make(rule.RuleListeners)

		// Variable declarations
		listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
			varDecl := node.AsVariableDeclaration()
			nameNode := varDecl.Name()
			if nameNode == nil {
				return
			}

			if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
				processBindingPattern(nameNode, report, node, ctx, exportedViaBlock)
				return
			}

			if !ast.IsIdentifier(nameNode) {
				return
			}

			name := nameNode.AsIdentifier().Text
			modifiers := detectVariableModifiers(node, ctx, exportedViaBlock)
			modifiers |= unusedModifier(nameNode)
			report(nameNode, name, SelectorVariable, modifiers)
		}

		// Function declarations
		listeners[ast.KindFunctionDeclaration] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}
			if !ast.IsIdentifier(nameNode) {
				return
			}
			name := nameNode.AsIdentifier().Text
			modifiers := detectFunctionModifiers(node, ctx, exportedViaBlock)
			modifiers |= unusedModifier(nameNode)
			report(nameNode, name, SelectorFunction, modifiers)
		}

		// Function expressions (named)
		listeners[ast.KindFunctionExpression] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}
			if !ast.IsIdentifier(nameNode) {
				return
			}
			name := nameNode.AsIdentifier().Text
			var modifiers Modifier
			if node.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
				modifiers |= ModifierAsync
			}
			report(nameNode, name, SelectorFunction, modifiers)
		}

		// Parameters
		listeners[ast.KindParameter] = func(node *ast.Node) {
			paramDecl := node.AsParameterDeclaration()
			nameNode := paramDecl.Name()
			if nameNode == nil {
				return
			}

			isParamProp := paramDecl.ModifierFlags()&ast.ModifierFlagsParameterPropertyModifier != 0

			if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
				if !isParamProp {
					processBindingPattern(nameNode, report, node, ctx, exportedViaBlock)
				}
				return
			}

			if !ast.IsIdentifier(nameNode) {
				return
			}

			name := nameNode.AsIdentifier().Text
			sel := SelectorParameter
			var modifiers Modifier
			if isParamProp {
				sel = SelectorParameterProperty
				modifiers = detectParameterPropertyModifiers(node)
			} else {
				modifiers = detectParameterModifiers(node)
				modifiers |= unusedModifier(nameNode)
			}
			report(nameNode, name, sel, modifiers)
		}

		// Class declarations
		listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
			modifiers := detectClassModifiers(node, ctx, exportedViaBlock)
			if nameNode := node.Name(); nameNode != nil {
				modifiers |= unusedModifier(nameNode)
			}
			processNode(node, SelectorClass, modifiers)
		}

		// Class expressions
		listeners[ast.KindClassExpression] = func(node *ast.Node) {
			processNode(node, SelectorClass, detectClassModifiers(node, ctx, exportedViaBlock))
		}

		// Interface declarations
		listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
			modifiers := detectExportedModifier(node, ctx, exportedViaBlock)
			if nameNode := node.Name(); nameNode != nil {
				modifiers |= unusedModifier(nameNode)
			}
			processNode(node, SelectorInterface, modifiers)
		}

		// Type alias declarations
		listeners[ast.KindTypeAliasDeclaration] = func(node *ast.Node) {
			modifiers := detectExportedModifier(node, ctx, exportedViaBlock)
			if nameNode := node.Name(); nameNode != nil {
				modifiers |= unusedModifier(nameNode)
			}
			processNode(node, SelectorTypeAlias, modifiers)
		}

		// Enum declarations
		listeners[ast.KindEnumDeclaration] = func(node *ast.Node) {
			modifiers := detectExportedModifier(node, ctx, exportedViaBlock)
			if node.ModifierFlags()&ast.ModifierFlagsConst != 0 {
				modifiers |= ModifierConst
			}
			if nameNode := node.Name(); nameNode != nil {
				modifiers |= unusedModifier(nameNode)
			}
			processNode(node, SelectorEnum, modifiers)
		}

		// Enum members
		listeners[ast.KindEnumMember] = func(node *ast.Node) {
			enumMember := node.AsEnumMember()
			nameNode := enumMember.Name()
			if nameNode == nil {
				return
			}

			var name string
			var modifiers Modifier
			modifiers |= ModifierPublic
			isStringLiteralName := false
			if ast.IsIdentifier(nameNode) {
				name = nameNode.AsIdentifier().Text
			} else if ast.IsStringLiteral(nameNode) {
				name = nameNode.AsStringLiteral().Text
				modifiers |= ModifierRequiresQuotes
				isStringLiteralName = true
			} else if ast.IsComputedPropertyName(nameNode) {
				return
			} else {
				name = nameNode.Text()
			}

			if name == "" && !isStringLiteralName {
				return
			}
			report(nameNode, name, SelectorEnumMember, modifiers)
		}

		// Type parameters
		listeners[ast.KindTypeParameter] = func(node *ast.Node) {
			var modifiers Modifier
			if nameNode := node.Name(); nameNode != nil {
				modifiers = unusedModifier(nameNode)
			}
			processNode(node, SelectorTypeParameter, modifiers)
		}

		// Property declarations (class properties and auto accessors)
		listeners[ast.KindPropertyDeclaration] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}

			sel := SelectorClassProperty
			if ast.IsAutoAccessorPropertyDeclaration(node) {
				sel = SelectorAutoAccessor
			}

			// Properties with function expression or arrow function values are classified as methods
			propDecl := node.AsPropertyDeclaration()
			if sel == SelectorClassProperty && propDecl.Initializer != nil && ast.IsFunctionExpressionOrArrowFunction(propDecl.Initializer) {
				sel = SelectorClassMethod
				modifiers := detectClassMethodModifiersFromProperty(node, propDecl.Initializer)
				processClassMember(nameNode, sel, modifiers, report)
				return
			}

			modifiers := detectPropertyModifiers(node)
			processClassMember(nameNode, sel, modifiers, report)
		}

		// Method declarations
		listeners[ast.KindMethodDeclaration] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}

			sel := detectMethodSelector(node)
			modifiers := detectMethodModifiers(node, sel)
			processClassMember(nameNode, sel, modifiers, report)
		}

		// Get/Set accessors
		listeners[ast.KindGetAccessor] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}
			modifiers := detectAccessorModifiers(node)
			processClassMember(nameNode, SelectorClassicAccessor, modifiers, report)
		}

		listeners[ast.KindSetAccessor] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}
			modifiers := detectAccessorModifiers(node)
			processClassMember(nameNode, SelectorClassicAccessor, modifiers, report)
		}

		// Property assignments (object literal properties)
		listeners[ast.KindPropertyAssignment] = func(node *ast.Node) {
			propAssign := node.AsPropertyAssignment()
			nameNode := propAssign.Name()
			if nameNode == nil {
				return
			}

			if ast.IsComputedPropertyName(nameNode) {
				return
			}

			// Properties with function expression or arrow function values are classified as methods
			sel := SelectorObjectLiteralProperty
			if propAssign.Initializer != nil && ast.IsFunctionExpressionOrArrowFunction(propAssign.Initializer) {
				sel = SelectorObjectLiteralMethod
			}

			var name string
			var modifiers Modifier
			modifiers |= ModifierPublic
			isStringLiteralName := false

			if ast.IsIdentifier(nameNode) {
				name = nameNode.AsIdentifier().Text
			} else if ast.IsStringLiteral(nameNode) {
				name = nameNode.AsStringLiteral().Text
				modifiers |= ModifierRequiresQuotes
				isStringLiteralName = true
			} else if ast.IsNumericLiteral(nameNode) {
				modifiers |= ModifierRequiresQuotes
				return
			} else {
				name = nameNode.Text()
			}

			if name == "" && !isStringLiteralName {
				return
			}

			// Add async modifier if the function value is async
			if sel == SelectorObjectLiteralMethod && propAssign.Initializer != nil &&
				propAssign.Initializer.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
				modifiers |= ModifierAsync
			}

			report(nameNode, name, sel, modifiers)
		}

		// Property signatures (type properties)
		listeners[ast.KindPropertySignature] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}

			if ast.IsComputedPropertyName(nameNode) {
				return
			}

			var name string
			var modifiers Modifier
			modifiers |= ModifierPublic
			isStringLiteralName := false

			if node.ModifierFlags()&ast.ModifierFlagsReadonly != 0 {
				modifiers |= ModifierReadonly
			}

			if ast.IsIdentifier(nameNode) {
				name = nameNode.AsIdentifier().Text
			} else if ast.IsStringLiteral(nameNode) {
				name = nameNode.AsStringLiteral().Text
				modifiers |= ModifierRequiresQuotes
				isStringLiteralName = true
			} else {
				name = nameNode.Text()
			}

			if name == "" && !isStringLiteralName {
				return
			}

			// Properties with function type annotations are classified as methods
			sel := SelectorTypeProperty
			propSig := node.AsPropertySignatureDeclaration()
			if propSig.Type != nil && propSig.Type.Kind == ast.KindFunctionType {
				sel = SelectorTypeMethod
			}

			report(nameNode, name, sel, modifiers)
		}

		// Method signatures (type methods)
		listeners[ast.KindMethodSignature] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}

			if ast.IsComputedPropertyName(nameNode) {
				return
			}

			var name string
			var modifiers Modifier
			modifiers |= ModifierPublic
			isStringLiteralName := false

			if ast.IsIdentifier(nameNode) {
				name = nameNode.AsIdentifier().Text
			} else if ast.IsStringLiteral(nameNode) {
				name = nameNode.AsStringLiteral().Text
				modifiers |= ModifierRequiresQuotes
				isStringLiteralName = true
			} else {
				name = nameNode.Text()
			}

			if name == "" && !isStringLiteralName {
				return
			}
			report(nameNode, name, SelectorTypeMethod, modifiers)
		}

		// Import clause (default imports)
		listeners[ast.KindImportClause] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}
			if !ast.IsIdentifier(nameNode) {
				return
			}
			name := nameNode.AsIdentifier().Text
			report(nameNode, name, SelectorImport, ModifierDefaultImport)
		}

		// Namespace imports
		listeners[ast.KindNamespaceImport] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}
			if !ast.IsIdentifier(nameNode) {
				return
			}
			name := nameNode.AsIdentifier().Text
			report(nameNode, name, SelectorImport, ModifierNamespace)
		}

		// Import specifiers (named imports with rename, including default destructured and string literal imports)
		listeners[ast.KindImportSpecifier] = func(node *ast.Node) {
			importSpec := node.AsImportSpecifier()
			if importSpec.PropertyName == nil {
				return // Named import without rename — not matched by import selector
			}
			propName := importSpec.PropertyName
			nameNode := importSpec.Name()
			if nameNode == nil || !ast.IsIdentifier(nameNode) {
				return
			}
			localName := nameNode.AsIdentifier().Text

			var modifiers Modifier
			if ast.IsIdentifier(propName) && propName.AsIdentifier().Text == "default" {
				modifiers = ModifierDefaultImport
			} else if ast.IsStringLiteral(propName) {
				// String literal import like import { "🍎" as Foo }
				// No special modifier, just SelectorImport
			} else {
				return // Regular renamed import { foo as bar } — not matched
			}
			report(nameNode, localName, SelectorImport, modifiers)
		}

		return listeners
	},
}

func parseOptions(options any) []NamingConventionOption {
	if options == nil {
		return defaultOptions
	}
	selectorOptions := utils.UnmarshalOptions[[]NamingConventionOption](options, "naming-convention")
	if len(selectorOptions) == 0 {
		return defaultOptions
	}
	return selectorOptions
}

func processBindingPattern(pattern *ast.Node, report func(*ast.Node, string, Selector, Modifier), parentDecl *ast.Node, ctx rule.RuleContext, exportedViaBlock map[string]bool) {
	bindingPattern := pattern.AsBindingPattern()
	if bindingPattern.Elements == nil {
		return
	}

	isParam := ast.IsParameter(parentDecl)

	for _, child := range bindingPattern.Elements.Nodes {
		if !ast.IsBindingElement(child) {
			continue
		}
		nameNode := child.Name()
		if nameNode == nil {
			continue
		}
		if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			processBindingPattern(nameNode, report, parentDecl, ctx, exportedViaBlock)
			continue
		}
		if ast.IsIdentifier(nameNode) {
			name := nameNode.AsIdentifier().Text
			// Only mark as destructured if the binding element has no property name
			// (shorthand like { name }). Renamed bindings like { prop: name } are
			// not considered destructured because the user chose the local name.
			isDestructured := child.AsBindingElement().PropertyName == nil
			if isParam {
				var modifiers Modifier
				if isDestructured {
					modifiers |= ModifierDestructured
				}
				report(nameNode, name, SelectorParameter, modifiers)
			} else {
				modifiers := detectVariableModifiers(parentDecl, ctx, exportedViaBlock)
				if isDestructured {
					modifiers |= ModifierDestructured
				}
				report(nameNode, name, SelectorVariable, modifiers)
			}
		}
	}
}

func processClassMember(nameNode *ast.Node, sel Selector, modifiers Modifier, report func(*ast.Node, string, Selector, Modifier)) {
	if ast.IsComputedPropertyName(nameNode) {
		return
	}

	var name string
	isStringLiteralName := false
	if ast.IsIdentifier(nameNode) {
		name = nameNode.AsIdentifier().Text
	} else if ast.IsPrivateIdentifier(nameNode) {
		name = nameNode.Text()
		modifiers |= ModifierHashPrivate
		// Strip the leading # for format validation
		if len(name) > 0 && name[0] == '#' {
			name = name[1:]
		}
	} else if ast.IsStringLiteral(nameNode) {
		name = nameNode.AsStringLiteral().Text
		modifiers |= ModifierRequiresQuotes
		isStringLiteralName = true
	} else {
		name = nameNode.Text()
	}

	if name == "" && !isStringLiteralName {
		return
	}
	report(nameNode, name, sel, modifiers)
}

// Modifier detection functions

func detectVariableModifiers(node *ast.Node, ctx rule.RuleContext, exportedViaBlock map[string]bool) Modifier {
	var modifiers Modifier

	if node.Parent != nil && ast.IsVariableDeclarationList(node.Parent) && node.Parent.Flags&ast.NodeFlagsConst != 0 {
		modifiers |= ModifierConst
	}

	if node.ModifierFlags()&ast.ModifierFlagsExport != 0 {
		modifiers |= ModifierExported
	} else if isExportedViaParent(node, exportedViaBlock) {
		modifiers |= ModifierExported
	}

	if isGlobalScope(node, ctx) {
		modifiers |= ModifierGlobal
	}

	varDecl := node.AsVariableDeclaration()
	if varDecl.Initializer != nil {
		init := varDecl.Initializer
		if (ast.IsArrowFunction(init) || ast.IsFunctionExpression(init)) && init.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
			modifiers |= ModifierAsync
		}
	}

	return modifiers
}

func detectFunctionModifiers(node *ast.Node, ctx rule.RuleContext, exportedViaBlock map[string]bool) Modifier {
	var modifiers Modifier

	if node.ModifierFlags()&ast.ModifierFlagsExport != 0 {
		modifiers |= ModifierExported
	} else if isExportedViaParent(node, exportedViaBlock) {
		modifiers |= ModifierExported
	}

	if isGlobalScope(node, ctx) {
		modifiers |= ModifierGlobal
	}

	if node.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
		modifiers |= ModifierAsync
	}

	return modifiers
}

func detectParameterModifiers(node *ast.Node) Modifier {
	var modifiers Modifier

	paramDecl := node.AsParameterDeclaration()
	nameNode := paramDecl.Name()
	if nameNode != nil && (ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode)) {
		modifiers |= ModifierDestructured
	}

	return modifiers
}

func detectParameterPropertyModifiers(node *ast.Node) Modifier {
	var modifiers Modifier
	flags := node.ModifierFlags()

	if flags&ast.ModifierFlagsReadonly != 0 {
		modifiers |= ModifierReadonly
	}
	modifiers |= detectAccessibility(flags)

	return modifiers
}

func detectClassModifiers(node *ast.Node, ctx rule.RuleContext, exportedViaBlock map[string]bool) Modifier {
	var modifiers Modifier
	flags := node.ModifierFlags()

	if flags&ast.ModifierFlagsAbstract != 0 {
		modifiers |= ModifierAbstract
	}

	if flags&ast.ModifierFlagsExport != 0 {
		modifiers |= ModifierExported
	} else if isExportedViaParent(node, exportedViaBlock) {
		modifiers |= ModifierExported
	}

	return modifiers
}

func detectExportedModifier(node *ast.Node, ctx rule.RuleContext, exportedViaBlock map[string]bool) Modifier {
	var modifiers Modifier

	if node.ModifierFlags()&ast.ModifierFlagsExport != 0 {
		modifiers |= ModifierExported
	} else if isExportedViaParent(node, exportedViaBlock) {
		modifiers |= ModifierExported
	}

	return modifiers
}

func detectPropertyModifiers(node *ast.Node) Modifier {
	var modifiers Modifier
	flags := node.ModifierFlags()

	if flags&ast.ModifierFlagsStatic != 0 {
		modifiers |= ModifierStatic
	}
	if flags&ast.ModifierFlagsReadonly != 0 {
		modifiers |= ModifierReadonly
	}
	if flags&ast.ModifierFlagsAbstract != 0 {
		modifiers |= ModifierAbstract
	}
	if flags&ast.ModifierFlagsOverride != 0 {
		modifiers |= ModifierOverride
	}

	modifiers |= detectAccessibility(flags)

	return modifiers
}

func detectMethodSelector(node *ast.Node) Selector {
	if node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindObjectLiteralExpression:
			return SelectorObjectLiteralMethod
		case ast.KindInterfaceDeclaration:
			return SelectorTypeMethod
		case ast.KindTypeLiteral:
			return SelectorTypeMethod
		}
	}
	return SelectorClassMethod
}

func detectMethodModifiers(node *ast.Node, sel Selector) Modifier {
	var modifiers Modifier
	flags := node.ModifierFlags()

	if sel == SelectorClassMethod {
		if flags&ast.ModifierFlagsStatic != 0 {
			modifiers |= ModifierStatic
		}
		if flags&ast.ModifierFlagsAbstract != 0 {
			modifiers |= ModifierAbstract
		}
		if flags&ast.ModifierFlagsOverride != 0 {
			modifiers |= ModifierOverride
		}
		modifiers |= detectAccessibility(flags)
	} else if sel == SelectorObjectLiteralMethod {
		modifiers |= ModifierPublic
	} else if sel == SelectorTypeMethod {
		modifiers |= ModifierPublic
	}

	if flags&ast.ModifierFlagsAsync != 0 {
		modifiers |= ModifierAsync
	}

	return modifiers
}

func detectAccessorModifiers(node *ast.Node) Modifier {
	var modifiers Modifier
	flags := node.ModifierFlags()

	if flags&ast.ModifierFlagsStatic != 0 {
		modifiers |= ModifierStatic
	}
	if flags&ast.ModifierFlagsAbstract != 0 {
		modifiers |= ModifierAbstract
	}
	if flags&ast.ModifierFlagsOverride != 0 {
		modifiers |= ModifierOverride
	}

	modifiers |= detectAccessibility(flags)

	return modifiers
}

// detectClassMethodModifiersFromProperty detects method modifiers for a class property
// that has been reclassified as a method (because its initializer is a function/arrow expression).
func detectClassMethodModifiersFromProperty(node *ast.Node, initializer *ast.Node) Modifier {
	var modifiers Modifier
	flags := node.ModifierFlags()

	if flags&ast.ModifierFlagsStatic != 0 {
		modifiers |= ModifierStatic
	}
	if flags&ast.ModifierFlagsAbstract != 0 {
		modifiers |= ModifierAbstract
	}
	if flags&ast.ModifierFlagsOverride != 0 {
		modifiers |= ModifierOverride
	}
	modifiers |= detectAccessibility(flags)

	// Async is on the initializer (the arrow/function expression), not on the property itself
	if initializer != nil && initializer.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
		modifiers |= ModifierAsync
	}

	return modifiers
}

func detectAccessibility(flags ast.ModifierFlags) Modifier {
	if flags&ast.ModifierFlagsPrivate != 0 {
		return ModifierPrivate
	}
	if flags&ast.ModifierFlagsProtected != 0 {
		return ModifierProtected
	}
	return ModifierPublic
}

func detectUnusedModifier(ctx rule.RuleContext, nameNode *ast.Node, exportedViaBlock map[string]bool) Modifier {
	if ctx.TypeChecker == nil || nameNode == nil {
		return 0
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if symbol == nil {
		return 0
	}
	// If the symbol is referenced in code, it's not unused
	if checker.Checker_isReferenced(ctx.TypeChecker, symbol) {
		return 0
	}
	// Exported symbols are considered "used" even if not referenced within the file
	if isDeclarationExported(nameNode, exportedViaBlock) {
		return 0
	}
	return ModifierUnused
}

// isDeclarationExported checks if the declaration containing nameNode has an export modifier
// or is exported via an `export { ... }` block.
func isDeclarationExported(nameNode *ast.Node, exportedViaBlock map[string]bool) bool {
	// Check export-via-block map
	if ast.IsIdentifier(nameNode) {
		name := nameNode.AsIdentifier().Text
		if exportedViaBlock[name] {
			return true
		}
	}
	node := nameNode.Parent
	if node == nil {
		return false
	}
	// For variable declarations, walk up: VarDecl → VDL → VarStatement
	if ast.IsVariableDeclaration(node) {
		node = node.Parent // VariableDeclarationList
		if node != nil {
			node = node.Parent // VariableStatement
		}
	}
	if node == nil {
		return false
	}
	return node.ModifierFlags()&ast.ModifierFlagsExport != 0
}

func detectTypeModifiers(ctx rule.RuleContext, node *ast.Node) TypeModifier {
	t := ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return 0
	}

	var result TypeModifier

	allParts := utils.UnionTypeParts(t)
	for _, part := range allParts {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBooleanLike) {
			result |= TypeModifierBoolean
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsStringLike) {
			result |= TypeModifierString
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLike) {
			result |= TypeModifierNumber
		}
		if len(utils.GetCallSignatures(ctx.TypeChecker, part)) > 0 {
			result |= TypeModifierFunction
		}
		if checker.Checker_isArrayOrTupleType(ctx.TypeChecker, part) {
			result |= TypeModifierArray
		}
	}

	return result
}

func isExportedViaParent(node *ast.Node, exportedViaBlock map[string]bool) bool {
	if node.Parent == nil {
		return false
	}
	parent := node.Parent
	if ast.IsVariableDeclarationList(parent) {
		parent = parent.Parent
	}
	if parent != nil && parent.ModifierFlags()&ast.ModifierFlagsExport != 0 {
		return true
	}
	// Check if exported via `export { name }` block
	if nameNode := node.Name(); nameNode != nil && ast.IsIdentifier(nameNode) {
		if exportedViaBlock[nameNode.AsIdentifier().Text] {
			return true
		}
	}
	return false
}

func buildExportedViaBlockMap(sourceFile *ast.SourceFile) map[string]bool {
	result := make(map[string]bool)
	if sourceFile.Statements == nil {
		return result
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if stmt.Kind != ast.KindExportDeclaration {
			continue
		}
		exportDecl := stmt.AsExportDeclaration()
		if exportDecl.ExportClause == nil {
			continue
		}
		if !ast.IsNamedExports(exportDecl.ExportClause) {
			continue
		}
		namedExports := exportDecl.ExportClause.AsNamedExports()
		if namedExports.Elements == nil {
			continue
		}
		for _, spec := range namedExports.Elements.Nodes {
			exportSpec := spec.AsExportSpecifier()
			// If PropertyName is set, the local name is PropertyName; export name is Name
			// If PropertyName is not set, both are the same (Name)
			var localName string
			if exportSpec.PropertyName != nil && ast.IsIdentifier(exportSpec.PropertyName) {
				localName = exportSpec.PropertyName.AsIdentifier().Text
			} else if exportSpec.PropertyName == nil {
				nameNode := exportSpec.Name()
				if nameNode != nil && ast.IsIdentifier(nameNode) {
					localName = nameNode.AsIdentifier().Text
				}
			}
			if localName != "" {
				result[localName] = true
			}
		}
	}
	return result
}

func isGlobalScope(node *ast.Node, ctx rule.RuleContext) bool {
	ancestor := node.Parent
	// For variables: VarDecl → VDL → VarStmt → SourceFile (unwrap twice)
	if ancestor != nil && ast.IsVariableDeclarationList(ancestor) {
		ancestor = ancestor.Parent // VariableStatement
		if ancestor == nil {
			return false
		}
		ancestor = ancestor.Parent // Should be SourceFile
	}
	if ancestor == nil {
		return false
	}
	// For functions: FuncDecl.Parent should be SourceFile directly.
	// For block-scoped: FuncDecl.Parent is Block → not SourceFile → false.
	if ancestor.Kind == ast.KindSourceFile {
		return !ast.IsExternalOrCommonJSModule(ctx.SourceFile)
	}
	return false
}
