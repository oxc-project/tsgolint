package naming_convention

import (
	"context"
	"fmt"
	"math/bits"
	"slices"
	"strings"
	"unicode"

	"github.com/dlclark/regexp2/v2"
	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

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
	// originalSelector is the selector as configured (possibly a meta
	// selector), before expansion; with modifierWeight it drives the config
	// sort (see compareSelectors).
	originalSelector Selector
	modifierWeight   int
	// selectorName is the human-readable selector name used in messages
	// (e.g. "Class Method"), precomputed once per selector config.
	selectorName string
}

// selectorGroups provides O(1) lookup by individual selector, indexed by bit position.
type selectorGroups [][]normalizedSelector

func selectorIndex(sel Selector) int {
	return bits.TrailingZeros(uint(sel))
}

// normalizedFilter is a compiled filter regex.
type normalizedFilter struct {
	regex *regexp2.Regexp
	match bool
}

// normalizedMatchRegex is a compiled custom regex.
type normalizedMatchRegex struct {
	regex *regexp2.Regexp
	match bool
}

// Message builder functions

func buildUnexpectedUnderscoreMessage(selectorName, name, position string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedUnderscore",
		Description: fmt.Sprintf("%s name `%s` must not have a %s underscore.", selectorName, name, position),
	}
}

func buildMissingUnderscoreMessage(selectorName, name, position, count string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingUnderscore",
		Description: fmt.Sprintf("%s name `%s` must have %s %s underscore(s).", selectorName, name, count, position),
	}
}

func buildMissingAffixMessage(selectorName, name, affixType string, affixes []string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAffix",
		Description: fmt.Sprintf("%s name `%s` must have one of the following %ses: %s", selectorName, name, affixType, strings.Join(affixes, ", ")),
	}
}

func buildSatisfyCustomMessage(selectorName, name, matchStr, regex string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "satisfyCustom",
		Description: fmt.Sprintf("%s name `%s` must %s the RegExp: %s", selectorName, name, matchStr, regex),
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

// selectorTypeToMessageString mirrors the upstream helper of the same name:
// "classMethod" -> "Class Method", "variable" -> "Variable".
func selectorTypeToMessageString(selectorType string) string {
	var sb strings.Builder
	for _, r := range selectorType {
		if unicode.IsUpper(r) {
			sb.WriteByte(' ')
		}
		sb.WriteRune(r)
	}
	notCamelCase := sb.String()
	if notCamelCase == "" {
		return notCamelCase
	}
	return strings.ToUpper(notCamelCase[:1]) + notCamelCase[1:]
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

// The checkers mirror upstream format.ts: an empty name is valid, only the
// first character's case-fold is inspected (so `$`, digits, and caseless
// scripts pass), and underscores are the only forbidden interior characters.

func isCamelCase(name string) bool {
	if name == "" {
		return true
	}
	r := firstRune(name)
	return r == unicode.ToLower(r) && !strings.Contains(name, "_")
}

func isStrictCamelCase(name string) bool {
	if name == "" {
		return true
	}
	r := firstRune(name)
	return r == unicode.ToLower(r) && hasStrictCamelHumps(name, false)
}

func isPascalCase(name string) bool {
	if name == "" {
		return true
	}
	r := firstRune(name)
	return r == unicode.ToUpper(r) && !strings.Contains(name, "_")
}

func isStrictPascalCase(name string) bool {
	if name == "" {
		return true
	}
	r := firstRune(name)
	return r == unicode.ToUpper(r) && hasStrictCamelHumps(name, true)
}

func isSnakeCase(name string) bool {
	return name == "" || (name == strings.ToLower(name) && validateUnderscores(name))
}

func isUpperCase(name string) bool {
	return name == "" || (name == strings.ToUpper(name) && validateUnderscores(name))
}

func firstRune(name string) rune {
	for _, r := range name {
		return r
	}
	return 0
}

// hasStrictCamelHumps mirrors upstream: no leading or interior underscores and
// no consecutive uppercase humps; wasUpper is the case of the first character.
func hasStrictCamelHumps(name string, wasUpper bool) bool {
	if strings.HasPrefix(name, "_") {
		return false
	}
	first := true
	for _, r := range name {
		if first {
			first = false
			continue
		}
		if r == '_' {
			return false
		}
		isUpper := r == unicode.ToUpper(r) && r != unicode.ToLower(r)
		if wasUpper == isUpper {
			if wasUpper {
				return false
			}
		} else {
			wasUpper = isUpper
		}
	}
	return true
}

// validateUnderscores checks for leading, trailing, and adjacent underscores.
func validateUnderscores(name string) bool {
	if strings.HasPrefix(name, "_") {
		return false
	}
	wasUnderscore := false
	first := true
	for _, r := range name {
		if first {
			first = false
			continue
		}
		if r == '_' {
			if wasUnderscore {
				return false
			}
			wasUnderscore = true
		} else {
			wasUnderscore = false
		}
	}
	return !wasUnderscore
}

// Option normalization

func normalizeOptions(rawOptions []NamingConventionOption) selectorGroups {
	var all []normalizedSelector
	for _, opt := range rawOptions {
		selectors := parseSelectorNames(opt.Selector)
		modifiers := parseModifiers(opt.Modifiers)
		types := parseTypeModifiers(opt.Types)
		formats := parseFormats(opt.Format)
		leadingUnderscore := parseUnderscoreOption(opt.LeadingUnderscore)
		trailingUnderscore := parseUnderscoreOption(opt.TrailingUnderscore)
		filter := parseFilter(opt.Filter)
		custom := parseCustomRegex(opt.Custom)

		for _, sel := range selectors {
			// Upstream's modifierWeight is the OR of the modifier and
			// type-modifier bit VALUES (not a count); the Modifier/TypeModifier
			// constants share upstream's exact bit layout. A filter adds the
			// most weight.
			modifierWeight := int(modifiers) | int(types)
			if filter != nil {
				modifierWeight |= 1 << 30
			}
			for _, expandedSel := range expandSelector(sel) {
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
					originalSelector:   sel,
					modifierWeight:     modifierWeight,
					selectorName:       selectorTypeToMessageString(selectorTypeString[expandedSel]),
				})
			}
		}
	}

	// Sort most specific first, mirroring upstream's comparator.
	slices.SortStableFunc(all, compareSelectors)

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
		// Unknown selector names are dropped, consistent with the array path;
		// falling back to SelectorDefault would turn a typo into a catch-all.
		return nil
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
	}
	// Non-string selector values (number/bool/object) are dropped for the same
	// reason as unknown names above: falling back to SelectorDefault would turn
	// a malformed config into a catch-all matching every identifier.
	return nil
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

func parseFormats(formats *[]string) []PredefinedFormat {
	if formats == nil {
		return nil
	}
	var result []PredefinedFormat
	for _, f := range *formats {
		if pf, ok := formatNameMap[f]; ok {
			result = append(result, pf)
		}
	}
	return result
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

// mustCompileOption compiles a user-supplied regex in ECMAScript mode (via
// regexp2, like switch-exhaustiveness-check) so that JS-only constructs such
// as lookahead and backreferences — accepted by upstream's `new RegExp(re,
// 'u')` — work here too. It panics on a genuinely malformed pattern: silently
// dropping an uncompilable filter would make its selector apply to every
// name, and dropping a custom regex would skip the check entirely — both
// silently wrong results. Panicking matches how utils.UnmarshalOptions
// handles invalid options.
func mustCompileOption(kind, regexStr string) *regexp2.Regexp {
	compiled, err := regexp2.Compile(regexStr, regexp2.ECMAScript|regexp2.Unicode)
	if err != nil {
		panic(fmt.Sprintf("naming-convention: invalid %s regex %q: %v", kind, regexStr, err))
	}
	return compiled
}

func parseFilter(filterRaw any) *normalizedFilter {
	if filterRaw == nil {
		return nil
	}
	switch v := filterRaw.(type) {
	case string:
		return &normalizedFilter{regex: mustCompileOption("filter", v), match: true}
	case map[string]any:
		regexStr, _ := v["regex"].(string)
		match, ok := v["match"].(bool)
		if !ok {
			match = true
		}
		return &normalizedFilter{regex: mustCompileOption("filter", regexStr), match: match}
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
			return &normalizedFilter{regex: mustCompileOption("filter", s), match: true}
		}
		return &normalizedFilter{regex: mustCompileOption("filter", mr.Regex), match: mr.Match}
	}
}

func parseCustomRegex(custom *MatchRegex) *normalizedMatchRegex {
	if custom == nil {
		return nil
	}
	return &normalizedMatchRegex{regex: mustCompileOption("custom", custom.Regex), match: custom.Match}
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
	return bits.OnesCount(uint(sel)) > 1
}

// compareSelectors mirrors upstream's config sort (parse-options.ts): configs
// naming the SAME selector are ordered by modifier weight descending; between
// different selectors the modifier weight is ignored — individual selectors
// precede meta selectors, the method/property metas precede other metas
// (upstream backward compat; accessor is deliberately NOT in this tier), and
// remaining ties order by raw selector value descending, so memberLike
// precedes accessor and default (-1) sorts last.
func compareSelectors(a, b normalizedSelector) int {
	if a.originalSelector == b.originalSelector {
		return b.modifierWeight - a.modifierWeight
	}
	aMeta := isMetaSelector(a.originalSelector)
	bMeta := isMetaSelector(b.originalSelector)
	if aMeta != bMeta {
		if aMeta {
			return 1
		}
		return -1
	}
	aMethodOrProperty := a.originalSelector == SelectorMethod || a.originalSelector == SelectorProperty
	bMethodOrProperty := b.originalSelector == SelectorMethod || b.originalSelector == SelectorProperty
	if aMethodOrProperty != bMethodOrProperty {
		if aMethodOrProperty {
			return -1
		}
		return 1
	}
	return int(b.originalSelector) - int(a.originalSelector)
}

// Validation pipeline

// selectorsWithTypesMask is the set of selectors whose `types` constraint is
// honored, mirroring upstream's SelectorsAllowedToHaveTypes. autoAccessor is
// deliberately absent (as upstream omits it): a `types` constraint on an
// autoAccessor config is ignored and the config applies unconditionally.
const selectorsWithTypesMask Selector = SelectorVariable | SelectorParameter | SelectorParameterProperty |
	SelectorClassProperty | SelectorObjectLiteralProperty | SelectorTypeProperty |
	SelectorClassicAccessor

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
			matched, _ := sel.filter.regex.MatchString(name)
			if matched != sel.filter.match {
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
		return runFormatValidation(name, sel, nodeModifiers)
	}

	return nil
}

func runFormatValidation(name string, sel *normalizedSelector, nodeModifiers Modifier) *rule.RuleMessage {
	processedName := name
	selectorName := sel.selectorName

	if sel.leadingUnderscore != 0 {
		stripped, msg := validateUnderscore("leading", processedName, sel.leadingUnderscore, selectorName, name)
		if msg != nil {
			return msg
		}
		processedName = stripped
	}

	if sel.trailingUnderscore != 0 {
		stripped, msg := validateUnderscore("trailing", processedName, sel.trailingUnderscore, selectorName, name)
		if msg != nil {
			return msg
		}
		processedName = stripped
	}

	if len(sel.prefix) > 0 {
		stripped, msg := validateAffix("prefix", processedName, sel.prefix, selectorName, name)
		if msg != nil {
			return msg
		}
		processedName = stripped
	}

	if len(sel.suffix) > 0 {
		stripped, msg := validateAffix("suffix", processedName, sel.suffix, selectorName, name)
		if msg != nil {
			return msg
		}
		processedName = stripped
	}

	if sel.custom != nil {
		if msg := validateCustomRegex(processedName, sel.custom, selectorName, name); msg != nil {
			return msg
		}
	}

	if len(sel.formats) > 0 {
		if msg := validatePredefinedFormat(name, processedName, sel.formats, selectorName, nodeModifiers); msg != nil {
			return msg
		}
	}

	return nil
}

func validateUnderscore(position string, name string, option UnderscoreOption, selectorName string, originalName string) (string, *rule.RuleMessage) {
	var hasSingle, hasDouble bool
	var trimSingle, trimDouble string
	if position == "leading" {
		hasSingle = strings.HasPrefix(name, "_")
		hasDouble = strings.HasPrefix(name, "__")
		trimSingle = strings.TrimPrefix(name, "_")
		trimDouble = strings.TrimPrefix(name, "__")
	} else {
		hasSingle = strings.HasSuffix(name, "_")
		hasDouble = strings.HasSuffix(name, "__")
		trimSingle = strings.TrimSuffix(name, "_")
		trimDouble = strings.TrimSuffix(name, "__")
	}

	switch option {
	// ALLOW - no conditions as the user doesn't care if it's there or not
	case UnderscoreAllow:
		if hasSingle {
			return trimSingle, nil
		}
		return name, nil

	case UnderscoreAllowDouble:
		if hasDouble {
			return trimDouble, nil
		}
		return name, nil

	case UnderscoreAllowSingleOrDouble:
		if hasDouble {
			return trimDouble, nil
		}
		if hasSingle {
			return trimSingle, nil
		}
		return name, nil

	// FORBID
	case UnderscoreForbid:
		if hasSingle {
			msg := buildUnexpectedUnderscoreMessage(selectorName, originalName, position)
			return "", &msg
		}
		return name, nil

	// REQUIRE
	case UnderscoreRequire:
		if !hasSingle {
			msg := buildMissingUnderscoreMessage(selectorName, originalName, position, "one")
			return "", &msg
		}
		return trimSingle, nil

	case UnderscoreRequireDouble:
		if !hasDouble {
			msg := buildMissingUnderscoreMessage(selectorName, originalName, position, "two")
			return "", &msg
		}
		return trimDouble, nil
	}

	return name, nil
}

func validateAffix(affixType string, name string, affixes []string, selectorName string, originalName string) (string, *rule.RuleMessage) {
	for _, affix := range affixes {
		if affixType == "prefix" {
			if strings.HasPrefix(name, affix) {
				return name[len(affix):], nil
			}
		} else {
			if strings.HasSuffix(name, affix) {
				if affix == "" {
					// Mirror upstream's name.slice(0, -affix.length): for an
					// empty suffix, slice(0, -0) trims the whole name.
					return "", nil
				}
				return name[:len(name)-len(affix)], nil
			}
		}
	}
	msg := buildMissingAffixMessage(selectorName, originalName, affixType, affixes)
	return "", &msg
}

func validateCustomRegex(name string, custom *normalizedMatchRegex, selectorName string, originalName string) *rule.RuleMessage {
	matched, _ := custom.regex.MatchString(name)
	if matched != custom.match {
		matchStr := "match"
		if !custom.match {
			matchStr = "not match"
		}
		msg := buildSatisfyCustomMessage(selectorName, originalName, matchStr, "/"+custom.regex.String()+"/u")
		return &msg
	}
	return nil
}

func validatePredefinedFormat(originalName string, processedName string, formats []PredefinedFormat, selectorName string, nodeModifiers Modifier) *rule.RuleMessage {
	// Upstream skips the format check for names that require quoting: such
	// names can never satisfy a format, so they are always reported.
	if nodeModifiers&ModifierRequiresQuotes == 0 {
		for _, format := range formats {
			if checkFormat(format, processedName) {
				return nil
			}
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

// nameRequiresQuotes mirrors upstream requiresQuoting: a member name written
// as a string literal only carries the requiresQuotes modifier when it is not
// a valid identifier.
func nameRequiresQuotes(name string) bool {
	return !scanner.IsIdentifierText(name, core.LanguageVariantStandard)
}

// extractDeclarationName returns the name text and quote-related modifiers for
// a declaration name node. Name kinds needing caller-specific handling
// (private identifiers) must be checked before calling.
// ok is false when the name is computed or empty.
func extractDeclarationName(nameNode *ast.Node) (string, Modifier, bool) {
	if ast.IsComputedPropertyName(nameNode) {
		return "", 0, false
	}
	var name string
	var modifiers Modifier
	isStringLiteralName := false
	switch {
	case ast.IsIdentifier(nameNode):
		name = nameNode.AsIdentifier().Text
	case ast.IsStringLiteral(nameNode):
		name = nameNode.AsStringLiteral().Text
		if nameRequiresQuotes(name) {
			modifiers |= ModifierRequiresQuotes
		}
		isStringLiteralName = true
	case ast.IsNumericLiteral(nameNode):
		// The scanner normalizes numeric literal text to its JS value string
		// ("0x10" -> "16"), matching upstream's `${node.value}`. Upstream then
		// applies the same requiresQuoting check as for string names — which a
		// numeric name always fails, so it is always reported when a format is
		// configured.
		name = nameNode.Text()
		if nameRequiresQuotes(name) {
			modifiers |= ModifierRequiresQuotes
		}
	default:
		name = nameNode.Text()
	}
	if name == "" && !isStringLiteralName {
		return "", 0, false
	}
	return name, modifiers, true
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

		// Force full type checking on this checker so that isReferenced returns
		// correct results. This is a deliberate tradeoff: it semantically checks
		// the whole file eagerly (work the linter otherwise does lazily), which
		// is why it only runs when a selector actually uses the `unused`
		// modifier. The port has no scope manager, so upstream's
		// variable.references bookkeeping has no cheaper equivalent here.
		// context.Background() is unavoidable: RuleContext carries no context.
		if needsUnusedCheck {
			checker.Checker_checkSourceFile(ctx.TypeChecker, context.Background(), ctx.SourceFile, true)
		}

		// Build a map of names exported via `export { ... }` blocks
		exportedViaBlock := buildExportedViaBlockMap(ctx.SourceFile)

		report := func(node *ast.Node, name string, nodeSelector Selector, nodeModifiers Modifier) {
			var nodeTypes TypeModifier
			// validateName only consults nodeTypes for selectors in
			// selectorsWithTypesMask, so skip the checker query for the rest.
			if needsTypeInfo && nodeSelector&selectorsWithTypesMask != 0 {
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

			name, mods, ok := extractDeclarationName(nameNode)
			if !ok {
				return
			}

			report(nameNode, name, nodeSelector, baseModifiers|mods)
		}

		listeners := make(rule.RuleListeners)

		// Variable declarations
		listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
			// typescript-go parses `catch (e)` as a VariableDeclaration, but
			// upstream never visits catch bindings.
			if node.Parent != nil && node.Parent.Kind == ast.KindCatchClause {
				return
			}

			varDecl := node.AsVariableDeclaration()
			nameNode := varDecl.Name()
			if nameNode == nil {
				return
			}

			if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
				processBindingPattern(nameNode, report, unusedModifier, node, exportedViaBlock)
				return
			}

			if !ast.IsIdentifier(nameNode) {
				return
			}

			name := nameNode.AsIdentifier().Text
			modifiers := detectVariableModifiers(node, exportedViaBlock)
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
			modifiers := detectFunctionModifiers(node, exportedViaBlock)
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

		// Parameters. Upstream only visits parameters of concrete function-like
		// declarations/expressions, never parameters in type positions
		// (function types, method/call/construct/index signatures).
		listeners[ast.KindParameter] = func(node *ast.Node) {
			if node.Parent == nil {
				return
			}
			switch node.Parent.Kind {
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
				ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			default:
				return
			}

			paramDecl := node.AsParameterDeclaration()
			nameNode := paramDecl.Name()
			if nameNode == nil {
				return
			}

			isParamProp := ast.IsParameterPropertyDeclaration(node, node.Parent)

			if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
				if !isParamProp {
					processBindingPattern(nameNode, report, unusedModifier, node, exportedViaBlock)
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
			modifiers := detectClassModifiers(node, exportedViaBlock)
			if nameNode := node.Name(); nameNode != nil {
				modifiers |= unusedModifier(nameNode)
			}
			processNode(node, SelectorClass, modifiers)
		}

		// Class expressions
		listeners[ast.KindClassExpression] = func(node *ast.Node) {
			processNode(node, SelectorClass, detectClassModifiers(node, exportedViaBlock))
		}

		// Interface declarations
		listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
			modifiers := detectExportedModifier(node, exportedViaBlock)
			if nameNode := node.Name(); nameNode != nil {
				modifiers |= unusedModifier(nameNode)
			}
			processNode(node, SelectorInterface, modifiers)
		}

		// Type alias declarations
		listeners[ast.KindTypeAliasDeclaration] = func(node *ast.Node) {
			modifiers := detectExportedModifier(node, exportedViaBlock)
			if nameNode := node.Name(); nameNode != nil {
				modifiers |= unusedModifier(nameNode)
			}
			processNode(node, SelectorTypeAlias, modifiers)
		}

		// Enum declarations. Upstream only ever assigns exported/unused here —
		// `const enum` does NOT get the const modifier.
		listeners[ast.KindEnumDeclaration] = func(node *ast.Node) {
			modifiers := detectExportedModifier(node, exportedViaBlock)
			if nameNode := node.Name(); nameNode != nil {
				modifiers |= unusedModifier(nameNode)
			}
			processNode(node, SelectorEnum, modifiers)
		}

		// Enum members. Upstream assigns them no accessibility modifier.
		listeners[ast.KindEnumMember] = func(node *ast.Node) {
			enumMember := node.AsEnumMember()
			nameNode := enumMember.Name()
			if nameNode == nil {
				return
			}

			name, modifiers, ok := extractDeclarationName(nameNode)
			if !ok {
				return
			}
			report(nameNode, name, SelectorEnumMember, modifiers)
		}

		// Type parameters
		listeners[ast.KindTypeParameter] = func(node *ast.Node) {
			// Upstream only selects `TSTypeParameterDeclaration > TSTypeParameter`,
			// i.e. type parameters that are members of a declaration's type
			// parameter list. The TS compiler AST also produces a KindTypeParameter
			// for the key of a mapped type (`{ [K in ...] }`) and the parameter of
			// an `infer` type, which upstream never visits, so skip those.
			if parent := node.Parent; parent != nil {
				switch parent.Kind {
				case ast.KindMappedType, ast.KindInferType:
					return
				}
			}
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

			name, modifiers, ok := extractDeclarationName(nameNode)
			if !ok {
				return
			}
			modifiers |= ModifierPublic

			// Add async modifier if the function value is async
			if sel == SelectorObjectLiteralMethod && propAssign.Initializer != nil &&
				propAssign.Initializer.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
				modifiers |= ModifierAsync
			}

			report(nameNode, name, sel, modifiers)
		}

		// Shorthand properties ({ foo }) are object literal properties too
		listeners[ast.KindShorthandPropertyAssignment] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil || !ast.IsIdentifier(nameNode) {
				return
			}
			report(nameNode, nameNode.AsIdentifier().Text, SelectorObjectLiteralProperty, ModifierPublic)
		}

		// Property signatures (type properties)
		listeners[ast.KindPropertySignature] = func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}

			name, modifiers, ok := extractDeclarationName(nameNode)
			if !ok {
				return
			}
			modifiers |= ModifierPublic
			if node.ModifierFlags()&ast.ModifierFlagsReadonly != 0 {
				modifiers |= ModifierReadonly
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

			name, modifiers, ok := extractDeclarationName(nameNode)
			if !ok {
				return
			}
			report(nameNode, name, SelectorTypeMethod, modifiers|ModifierPublic)
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
				// String literal import like import { "🍎" as Foo }: upstream's
				// guard only skips Identifier imported names other than
				// `default`, so any non-identifier name falls through and gets
				// the default modifier.
				modifiers = ModifierDefaultImport
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

func processBindingPattern(pattern *ast.Node, report func(*ast.Node, string, Selector, Modifier), unusedModifier func(*ast.Node) Modifier, parentDecl *ast.Node, exportedViaBlock map[string]bool) {
	bindingPattern := pattern.AsBindingPattern()
	if bindingPattern.Elements == nil {
		return
	}

	isParam := ast.IsParameterDeclaration(parentDecl)

	for _, child := range bindingPattern.Elements.Nodes {
		if !ast.IsBindingElement(child) {
			continue
		}
		nameNode := child.Name()
		if nameNode == nil {
			continue
		}
		if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			processBindingPattern(nameNode, report, unusedModifier, parentDecl, exportedViaBlock)
			continue
		}
		if ast.IsIdentifier(nameNode) {
			name := nameNode.AsIdentifier().Text
			// Only object-pattern shorthand like { name } is destructured
			// upstream. Renamed bindings like { prop: name } are not (the user
			// chose the local name), and array-pattern elements never are.
			isDestructured := ast.IsObjectBindingPattern(pattern) && child.AsBindingElement().PropertyName == nil
			if isParam {
				var modifiers Modifier
				if isDestructured {
					modifiers |= ModifierDestructured
				}
				modifiers |= unusedModifier(nameNode)
				report(nameNode, name, SelectorParameter, modifiers)
			} else {
				modifiers := detectVariableModifiers(parentDecl, exportedViaBlock)
				if isDestructured {
					modifiers |= ModifierDestructured
				}
				modifiers |= unusedModifier(nameNode)
				report(nameNode, name, SelectorVariable, modifiers)
			}
		}
	}
}

func processClassMember(nameNode *ast.Node, sel Selector, modifiers Modifier, report func(*ast.Node, string, Selector, Modifier)) {
	if ast.IsPrivateIdentifier(nameNode) {
		name := nameNode.Text()
		// Strip the leading # for format validation
		if len(name) > 0 && name[0] == '#' {
			name = name[1:]
		}
		if name == "" {
			return
		}
		// Upstream adds #private exclusively: a private-identifier member never
		// gets the public modifier (accessibility keywords are illegal on it,
		// so public is the only detectAccessibility result to strip).
		report(nameNode, name, sel, (modifiers&^ModifierPublic)|ModifierHashPrivate)
		return
	}

	name, nameModifiers, ok := extractDeclarationName(nameNode)
	if !ok {
		return
	}
	report(nameNode, name, sel, modifiers|nameModifiers)
}

// Modifier detection functions

func detectVariableModifiers(node *ast.Node, exportedViaBlock map[string]bool) Modifier {
	var modifiers Modifier

	if node.Parent != nil && ast.IsVariableDeclarationList(node.Parent) && node.Parent.Flags&ast.NodeFlagsConst != 0 {
		modifiers |= ModifierConst
	}

	modifiers |= detectExportedModifier(node, exportedViaBlock)

	if isGlobalScope(node) {
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

func detectFunctionModifiers(node *ast.Node, exportedViaBlock map[string]bool) Modifier {
	var modifiers Modifier

	modifiers |= detectExportedModifier(node, exportedViaBlock)

	if isGlobalScope(node) {
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

func detectClassModifiers(node *ast.Node, exportedViaBlock map[string]bool) Modifier {
	var modifiers Modifier
	flags := node.ModifierFlags()

	if flags&ast.ModifierFlagsAbstract != 0 {
		modifiers |= ModifierAbstract
	}

	modifiers |= detectExportedModifier(node, exportedViaBlock)

	return modifiers
}

func detectExportedModifier(node *ast.Node, exportedViaBlock map[string]bool) Modifier {
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
	if node.ModifierFlags()&ast.ModifierFlagsExport != 0 {
		return true
	}
	// `export { name }` re-exports module-scope bindings only, so consult the
	// name-keyed map just for top-level declarations; nested locals merely
	// sharing the name must not be treated as exported.
	if ast.IsIdentifier(nameNode) && node.Parent != nil && node.Parent.Kind == ast.KindSourceFile {
		return exportedViaBlock[nameNode.AsIdentifier().Text]
	}
	return false
}

// detectTypeModifiers mirrors upstream isCorrectType: strip null/undefined,
// then require EVERY union constituent to match for array/function, and exact
// whole-type string equality on the widened base literal type for
// boolean/string/number — so a mixed union like `string | number` matches
// no type modifier at all.
func detectTypeModifiers(ctx rule.RuleContext, node *ast.Node) TypeModifier {
	t := ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return 0
	}
	t = checker.Checker_GetNonNullableType(ctx.TypeChecker, t)

	var result TypeModifier

	parts := utils.UnionTypeParts(t)
	allPartsMatch := func(pred func(part *checker.Type) bool) bool {
		for _, part := range parts {
			if !pred(part) {
				return false
			}
		}
		return len(parts) > 0
	}

	if allPartsMatch(func(part *checker.Type) bool {
		return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, part)
	}) {
		result |= TypeModifierArray
	}
	if allPartsMatch(func(part *checker.Type) bool {
		return len(utils.GetCallSignatures(ctx.TypeChecker, part)) > 0
	}) {
		result |= TypeModifierFunction
	}

	widened := checker.Checker_getWidenedType(ctx.TypeChecker, checker.Checker_getBaseTypeOfLiteralType(ctx.TypeChecker, t))
	switch ctx.TypeChecker.TypeToString(widened) {
	case "boolean":
		result |= TypeModifierBoolean
	case "string":
		result |= TypeModifierString
	case "number":
		result |= TypeModifierNumber
	}

	return result
}

func isExportedViaParent(node *ast.Node, exportedViaBlock map[string]bool) bool {
	if node.Parent == nil {
		return false
	}
	stmt := node
	parent := node.Parent
	if ast.IsVariableDeclarationList(parent) {
		stmt = parent.Parent
		parent = parent.Parent
	}
	if parent != nil && parent.ModifierFlags()&ast.ModifierFlagsExport != 0 {
		return true
	}
	// `export { name }` re-exports module-scope bindings only, so consult the
	// name-keyed map just for top-level declarations; nested locals merely
	// sharing the name must not inherit the exported modifier.
	if stmt == nil || stmt.Parent == nil || stmt.Parent.Kind != ast.KindSourceFile {
		return false
	}
	if nameNode := node.Name(); nameNode != nil && ast.IsIdentifier(nameNode) {
		return exportedViaBlock[nameNode.AsIdentifier().Text]
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

func isGlobalScope(node *ast.Node) bool {
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
	return ast.IsGlobalSourceFile(ancestor)
}
