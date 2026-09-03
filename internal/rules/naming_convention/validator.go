package naming_convention

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/dlclark/regexp2/v2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// selectorFlags is a bit set of individual selectors. Meta selectors are unions
// of individual selectors, and the `default` selector is -1, mirroring
// typescript-eslint so that the sorting of selectors behaves identically.
type selectorFlags int64

const (
	// variableLike
	selectorVariable selectorFlags = 1 << iota
	selectorFunction
	selectorParameter

	// memberLike
	selectorParameterProperty
	selectorClassicAccessor
	selectorEnumMember
	selectorClassMethod
	selectorObjectLiteralMethod
	selectorTypeMethod
	selectorClassProperty
	selectorObjectLiteralProperty
	selectorTypeProperty
	selectorAutoAccessor

	// typeLike
	selectorClass
	selectorInterface
	selectorTypeAlias
	selectorEnum
	selectorTypeParameter

	// other
	selectorImport
)

const (
	metaSelectorDefault      selectorFlags = -1
	metaSelectorVariableLike               = selectorVariable | selectorFunction | selectorParameter
	metaSelectorMemberLike                 = selectorClassProperty | selectorObjectLiteralProperty | selectorTypeProperty | selectorParameterProperty | selectorEnumMember | selectorClassMethod | selectorObjectLiteralMethod | selectorTypeMethod | selectorClassicAccessor | selectorAutoAccessor
	metaSelectorTypeLike                   = selectorClass | selectorInterface | selectorTypeAlias | selectorEnum | selectorTypeParameter
	metaSelectorMethod                     = selectorClassMethod | selectorObjectLiteralMethod | selectorTypeMethod
	metaSelectorProperty                   = selectorClassProperty | selectorObjectLiteralProperty | selectorTypeProperty
	metaSelectorAccessor                   = selectorClassicAccessor | selectorAutoAccessor
)

var selectorsByName = map[string]selectorFlags{
	"default":      metaSelectorDefault,
	"variableLike": metaSelectorVariableLike,
	"memberLike":   metaSelectorMemberLike,
	"typeLike":     metaSelectorTypeLike,
	"method":       metaSelectorMethod,
	"property":     metaSelectorProperty,
	"accessor":     metaSelectorAccessor,

	"variable":              selectorVariable,
	"function":              selectorFunction,
	"parameter":             selectorParameter,
	"parameterProperty":     selectorParameterProperty,
	"classicAccessor":       selectorClassicAccessor,
	"enumMember":            selectorEnumMember,
	"classMethod":           selectorClassMethod,
	"objectLiteralMethod":   selectorObjectLiteralMethod,
	"typeMethod":            selectorTypeMethod,
	"classProperty":         selectorClassProperty,
	"objectLiteralProperty": selectorObjectLiteralProperty,
	"typeProperty":          selectorTypeProperty,
	"autoAccessor":          selectorAutoAccessor,
	"class":                 selectorClass,
	"interface":             selectorInterface,
	"typeAlias":             selectorTypeAlias,
	"enum":                  selectorEnum,
	"typeParameter":         selectorTypeParameter,
	"import":                selectorImport,
}

// individualSelectors lists every individual (non-meta) selector together with
// its name, in declaration order.
var individualSelectors = []struct {
	name  string
	value selectorFlags
}{
	{"variable", selectorVariable},
	{"function", selectorFunction},
	{"parameter", selectorParameter},
	{"parameterProperty", selectorParameterProperty},
	{"classicAccessor", selectorClassicAccessor},
	{"enumMember", selectorEnumMember},
	{"classMethod", selectorClassMethod},
	{"objectLiteralMethod", selectorObjectLiteralMethod},
	{"typeMethod", selectorTypeMethod},
	{"classProperty", selectorClassProperty},
	{"objectLiteralProperty", selectorObjectLiteralProperty},
	{"typeProperty", selectorTypeProperty},
	{"autoAccessor", selectorAutoAccessor},
	{"class", selectorClass},
	{"interface", selectorInterface},
	{"typeAlias", selectorTypeAlias},
	{"enum", selectorEnum},
	{"typeParameter", selectorTypeParameter},
	{"import", selectorImport},
}

func isMetaSelector(s selectorFlags) bool {
	switch s {
	case metaSelectorDefault, metaSelectorVariableLike, metaSelectorMemberLike, metaSelectorTypeLike, metaSelectorMethod, metaSelectorProperty, metaSelectorAccessor:
		return true
	}
	return false
}

func isMethodOrPropertySelector(s selectorFlags) bool {
	return s == metaSelectorMethod || s == metaSelectorProperty
}

// modifierFlags is a bit set of modifiers (and type modifiers). The bit
// positions mirror typescript-eslint so that the modifier weight used for
// ordering selectors is identical.
type modifierFlags int64

const (
	modifierConst modifierFlags = 1 << iota
	modifierReadonly
	modifierStatic
	modifierPublic
	modifierProtected
	modifierPrivate
	modifierHashPrivate
	modifierAbstract
	modifierDestructured
	modifierGlobal
	modifierExported
	modifierUnused
	modifierRequiresQuotes
	modifierOverride
	modifierAsync
	modifierDefault
	modifierNamespace

	// type modifiers start right after the modifiers so that sorting works
	typeModifierBoolean
	typeModifierString
	typeModifierNumber
	typeModifierFunction
	typeModifierArray
)

var modifiersByName = map[string]modifierFlags{
	"const":          modifierConst,
	"readonly":       modifierReadonly,
	"static":         modifierStatic,
	"public":         modifierPublic,
	"protected":      modifierProtected,
	"private":        modifierPrivate,
	"#private":       modifierHashPrivate,
	"abstract":       modifierAbstract,
	"destructured":   modifierDestructured,
	"global":         modifierGlobal,
	"exported":       modifierExported,
	"unused":         modifierUnused,
	"requiresQuotes": modifierRequiresQuotes,
	"override":       modifierOverride,
	"async":          modifierAsync,
	"default":        modifierDefault,
	"namespace":      modifierNamespace,
}

var typeModifiersByName = map[string]modifierFlags{
	"boolean":  typeModifierBoolean,
	"string":   typeModifierString,
	"number":   typeModifierNumber,
	"function": typeModifierFunction,
	"array":    typeModifierArray,
}

// selectors with a filter get the highest priority
const filterWeight modifierFlags = 1 << 30

type predefinedFormat uint8

const (
	formatCamelCase predefinedFormat = iota + 1
	formatStrictCamelCase
	formatPascalCase
	formatStrictPascalCase
	formatSnakeCase
	formatUpperCase
)

var formatsByName = map[string]predefinedFormat{
	"camelCase":        formatCamelCase,
	"strictCamelCase":  formatStrictCamelCase,
	"PascalCase":       formatPascalCase,
	"StrictPascalCase": formatStrictPascalCase,
	"snake_case":       formatSnakeCase,
	"UPPER_CASE":       formatUpperCase,
}

func (f predefinedFormat) String() string {
	switch f {
	case formatCamelCase:
		return "camelCase"
	case formatStrictCamelCase:
		return "strictCamelCase"
	case formatPascalCase:
		return "PascalCase"
	case formatStrictPascalCase:
		return "StrictPascalCase"
	case formatSnakeCase:
		return "snake_case"
	case formatUpperCase:
		return "UPPER_CASE"
	}
	return ""
}

func (f predefinedFormat) check(name string) bool {
	switch f {
	case formatCamelCase:
		return isCamelCase(name)
	case formatStrictCamelCase:
		return isStrictCamelCase(name)
	case formatPascalCase:
		return isPascalCase(name)
	case formatStrictPascalCase:
		return isStrictPascalCase(name)
	case formatSnakeCase:
		return isSnakeCase(name)
	case formatUpperCase:
		return isUpperCase(name)
	}
	return false
}

type underscoreOption uint8

const (
	underscoreUnset underscoreOption = iota
	underscoreForbid
	underscoreAllow
	underscoreRequire
	underscoreRequireDouble
	underscoreAllowDouble
	underscoreAllowSingleOrDouble
)

var underscoreOptionsByName = map[string]underscoreOption{
	"forbid":              underscoreForbid,
	"allow":               underscoreAllow,
	"require":             underscoreRequire,
	"requireDouble":       underscoreRequireDouble,
	"allowDouble":         underscoreAllowDouble,
	"allowSingleOrDouble": underscoreAllowSingleOrDouble,
}

type matchRegex struct {
	match  bool
	regex  *regexp2.Regexp
	source string
}

func (m *matchRegex) test(name string) bool {
	matched, _ := m.regex.MatchString(name)
	return matched
}

// normalizedSelector is a single selector option after parsing, with a single
// (individual or meta) selector.
type normalizedSelector struct {
	selector           selectorFlags
	modifiers          []modifierFlags
	types              []modifierFlags
	custom             *matchRegex
	filter             *matchRegex
	format             []predefinedFormat
	leadingUnderscore  underscoreOption
	trailingUnderscore underscoreOption
	prefix             []string
	suffix             []string
	// calculated ordering weight based on modifiers
	modifierWeight modifierFlags
}

// defaultOptions mirrors typescript-eslint's default config, which essentially
// mirrors ESLint's `camelcase` rule.
var defaultOptions = NamingConventionOptions{
	{
		Selector:           "default",
		Format:             []any{"camelCase"},
		LeadingUnderscore:  ptr(NamingConventionUnderscoreOptionAllow),
		TrailingUnderscore: ptr(NamingConventionUnderscoreOptionAllow),
	},
	{
		Selector: "import",
		Format:   []any{"camelCase", "PascalCase"},
	},
	{
		Selector:           "variable",
		Format:             []any{"camelCase", "UPPER_CASE"},
		LeadingUnderscore:  ptr(NamingConventionUnderscoreOptionAllow),
		TrailingUnderscore: ptr(NamingConventionUnderscoreOptionAllow),
	},
	{
		Selector: "typeLike",
		Format:   []any{"PascalCase"},
	},
}

func ptr[T any](v T) *T {
	return &v
}

func parseOptions(options any) NamingConventionOptions {
	// Be lenient and accept a single selector object instead of an array.
	if single, ok := options.(map[string]any); ok {
		options = []any{single}
	}
	opts := utils.UnmarshalOptions[NamingConventionOptions](options, "naming-convention")
	if len(opts) == 0 {
		// only apply the defaults when the user provides no config
		return defaultOptions
	}
	return opts
}

func invalidOptionf(format string, args ...any) {
	panic("naming-convention: " + fmt.Sprintf(format, args...))
}

func compileRegex(source string) *regexp2.Regexp {
	// typescript-eslint creates the RegExp with the `u` flag
	re, err := regexp2.Compile(source, regexp2.ECMAScript|regexp2.Unicode)
	if err != nil {
		invalidOptionf("invalid regular expression %q: %v", source, err)
	}
	return re
}

func parseStringOrStrings(value any, field string) []string {
	switch v := value.(type) {
	case string:
		return []string{v}
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				invalidOptionf("expected `%s` to be a string or an array of strings", field)
			}
			result = append(result, s)
		}
		return result
	}
	invalidOptionf("expected `%s` to be a string or an array of strings", field)
	return nil
}

func normalizeOption(option NamingConventionSelector) []normalizedSelector {
	var weight modifierFlags

	var modifiers []modifierFlags
	if option.Modifiers != nil {
		modifiers = make([]modifierFlags, 0, len(option.Modifiers))
		for _, name := range option.Modifiers {
			mod, ok := modifiersByName[string(name)]
			if !ok {
				invalidOptionf("unknown modifier %q", name)
			}
			modifiers = append(modifiers, mod)
			weight |= mod
		}
	}

	var types []modifierFlags
	if option.Types != nil {
		types = make([]modifierFlags, 0, len(option.Types))
		for _, name := range option.Types {
			mod, ok := typeModifiersByName[string(name)]
			if !ok {
				invalidOptionf("unknown type %q", name)
			}
			types = append(types, mod)
			weight |= mod
		}
	}

	var filter *matchRegex
	switch f := option.Filter.(type) {
	case nil:
	case string:
		filter = &matchRegex{match: true, regex: compileRegex(f), source: f}
	case map[string]any:
		match, matchOk := f["match"].(bool)
		source, regexOk := f["regex"].(string)
		if !matchOk || !regexOk {
			invalidOptionf("expected `filter` to be a string or an object with `match` and `regex` properties")
		}
		filter = &matchRegex{match: match, regex: compileRegex(source), source: source}
	default:
		invalidOptionf("expected `filter` to be a string or an object with `match` and `regex` properties")
	}
	if filter != nil {
		// give selectors with a filter the _highest_ priority
		weight |= filterWeight
	}

	var custom *matchRegex
	if option.Custom != nil {
		custom = &matchRegex{match: option.Custom.Match, regex: compileRegex(option.Custom.Regex), source: option.Custom.Regex}
	}

	var format []predefinedFormat
	if option.Format != nil {
		for _, name := range parseStringOrStrings(option.Format, "format") {
			f, ok := formatsByName[name]
			if !ok {
				invalidOptionf("unknown format %q", name)
			}
			format = append(format, f)
		}
	}

	var leadingUnderscore, trailingUnderscore underscoreOption
	if option.LeadingUnderscore != nil {
		leadingUnderscore = underscoreOptionsByName[string(*option.LeadingUnderscore)]
	}
	if option.TrailingUnderscore != nil {
		trailingUnderscore = underscoreOptionsByName[string(*option.TrailingUnderscore)]
	}

	normalized := normalizedSelector{
		custom:             custom,
		filter:             filter,
		format:             format,
		leadingUnderscore:  leadingUnderscore,
		trailingUnderscore: trailingUnderscore,
		modifiers:          modifiers,
		prefix:             option.Prefix,
		suffix:             option.Suffix,
		types:              types,
		modifierWeight:     weight,
	}

	selectorNames := parseStringOrStrings(option.Selector, "selector")
	result := make([]normalizedSelector, 0, len(selectorNames))
	for _, name := range selectorNames {
		selector, ok := selectorsByName[name]
		if !ok {
			invalidOptionf("unknown selector %q", name)
		}
		s := normalized
		s.selector = selector
		result = append(result, s)
	}
	return result
}

// validator validates the names of one individual selector.
type validator struct {
	selectorType selectorFlags
	// human readable selector name used in messages, e.g. "Class Property"
	typeName string
	// the applicable configs, sorted from highest to lowest priority
	configs []*normalizedSelector
}

func createValidators(options NamingConventionOptions) map[selectorFlags]*validator {
	var allConfigs []normalizedSelector
	for _, option := range options {
		allConfigs = append(allConfigs, normalizeOption(option)...)
	}

	validators := make(map[selectorFlags]*validator, len(individualSelectors))
	for _, individual := range individualSelectors {
		validators[individual.value] = createValidator(individual.name, individual.value, allConfigs)
	}
	return validators
}

func createValidator(name string, selectorType selectorFlags, allConfigs []normalizedSelector) *validator {
	// gather all of the applicable selectors
	configs := make([]*normalizedSelector, 0)
	for i := range allConfigs {
		c := &allConfigs[i]
		if c.selector&selectorType != 0 || c.selector == metaSelectorDefault {
			configs = append(configs, c)
		}
	}

	// make sure the "highest priority" configs are checked first
	slices.SortStableFunc(configs, func(a, b *normalizedSelector) int {
		if a.selector == b.selector {
			// in the event of the same selector, order by modifier weight
			// sort descending - the type modifiers are "more important"
			return cmp.Compare(b.modifierWeight, a.modifierWeight)
		}

		aIsMeta := isMetaSelector(a.selector)
		bIsMeta := isMetaSelector(b.selector)

		// non-meta selectors should go ahead of meta selectors
		if aIsMeta && !bIsMeta {
			return 1
		}
		if !aIsMeta && bIsMeta {
			return -1
		}

		aIsMethodOrProperty := isMethodOrPropertySelector(a.selector)
		bIsMethodOrProperty := isMethodOrPropertySelector(b.selector)

		// for backward compatibility, method and property have higher precedence than other meta selectors
		if aIsMethodOrProperty && !bIsMethodOrProperty {
			return -1
		}
		if !aIsMethodOrProperty && bIsMethodOrProperty {
			return 1
		}

		// both aren't meta selectors
		// sort descending - the meta selectors are "least important"
		return cmp.Compare(b.selector, a.selector)
	})

	return &validator{
		selectorType: selectorType,
		typeName:     selectorTypeToMessageString(name),
		configs:      configs,
	}
}

// selectorTypeToMessageString converts e.g. "classProperty" to "Class Property".
func selectorTypeToMessageString(selectorType string) string {
	var b strings.Builder
	for i, r := range selectorType {
		if i == 0 {
			b.WriteRune(unicode.ToUpper(r))
			continue
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

type nameValidationContext struct {
	ctx      rule.RuleContext
	typeName string
	// the range to report on
	reportRange  core.TextRange
	originalName string
}

// validate validates the name of the given name node (an identifier, private
// identifier or literal), reporting on reportRange.
func (v *validator) validate(ctx rule.RuleContext, node *ast.Node, reportRange core.TextRange, originalName string, modifiers modifierFlags) {
	if len(v.configs) == 0 {
		return
	}

	nctx := nameValidationContext{ctx: ctx, typeName: v.typeName, reportRange: reportRange, originalName: originalName}

	// return will break the loop and stop checking configs
	// it is only used when the name is known to have failed or succeeded a config.
	for _, config := range v.configs {
		if config.filter != nil && config.filter.test(originalName) != config.filter.match {
			// name does not match the filter
			continue
		}

		if slices.ContainsFunc(config.modifiers, func(m modifierFlags) bool { return modifiers&m == 0 }) {
			// does not have the required modifiers
			continue
		}

		if !isCorrectType(ctx, node, config, v.selectorType) {
			// is not the correct type
			continue
		}

		name := originalName
		var ok bool

		name, ok = nctx.validateUnderscore("leading", config, name)
		if !ok {
			return
		}

		name, ok = nctx.validateUnderscore("trailing", config, name)
		if !ok {
			return
		}

		name, ok = nctx.validateAffix("prefix", config, name)
		if !ok {
			return
		}

		name, ok = nctx.validateAffix("suffix", config, name)
		if !ok {
			return
		}

		if !nctx.validateCustom(config, name) {
			return
		}

		if !nctx.validatePredefinedFormat(config, name, modifiers) {
			return
		}

		// it's valid for this config, so we don't need to check any more configs
		return
	}
}

func (n *nameValidationContext) report(id string, description string) {
	n.ctx.ReportRange(n.reportRange, rule.RuleMessage{
		Id:          id,
		Description: description,
	})
}

// validateUnderscore returns the name with the underscore removed, if it is
// valid according to the specified underscore option, and false otherwise.
func (n *nameValidationContext) validateUnderscore(position string, config *normalizedSelector, name string) (string, bool) {
	option := config.leadingUnderscore
	if position == "trailing" {
		option = config.trailingUnderscore
	}
	if option == underscoreUnset {
		return name, true
	}

	leading := position == "leading"
	hasSingleUnderscore := func() bool {
		if leading {
			return strings.HasPrefix(name, "_")
		}
		return strings.HasSuffix(name, "_")
	}
	trimSingleUnderscore := func() string {
		if leading {
			return name[1:]
		}
		return name[:len(name)-1]
	}
	hasDoubleUnderscore := func() bool {
		if leading {
			return strings.HasPrefix(name, "__")
		}
		return strings.HasSuffix(name, "__")
	}
	trimDoubleUnderscore := func() string {
		if leading {
			return name[2:]
		}
		return name[:len(name)-2]
	}

	switch option {
	// ALLOW - no conditions as the user doesn't care if it's there or not
	case underscoreAllow:
		if hasSingleUnderscore() {
			return trimSingleUnderscore(), true
		}
		return name, true

	case underscoreAllowDouble:
		if hasDoubleUnderscore() {
			return trimDoubleUnderscore(), true
		}
		return name, true

	case underscoreAllowSingleOrDouble:
		if hasDoubleUnderscore() {
			return trimDoubleUnderscore(), true
		}
		if hasSingleUnderscore() {
			return trimSingleUnderscore(), true
		}
		return name, true

	// FORBID
	case underscoreForbid:
		if hasSingleUnderscore() {
			n.report("unexpectedUnderscore", fmt.Sprintf("%s name `%s` must not have a %s underscore.", n.typeName, n.originalName, position))
			return "", false
		}
		return name, true

	// REQUIRE
	case underscoreRequire:
		if !hasSingleUnderscore() {
			n.report("missingUnderscore", fmt.Sprintf("%s name `%s` must have one %s underscore(s).", n.typeName, n.originalName, position))
			return "", false
		}
		return trimSingleUnderscore(), true

	case underscoreRequireDouble:
		if !hasDoubleUnderscore() {
			n.report("missingUnderscore", fmt.Sprintf("%s name `%s` must have two %s underscore(s).", n.typeName, n.originalName, position))
			return "", false
		}
		return trimDoubleUnderscore(), true
	}

	return name, true
}

// validateAffix returns the name with the affix removed, if it is valid
// according to the specified affix option, and false otherwise.
func (n *nameValidationContext) validateAffix(position string, config *normalizedSelector, name string) (string, bool) {
	affixes := config.prefix
	if position == "suffix" {
		affixes = config.suffix
	}
	if len(affixes) == 0 {
		return name, true
	}

	for _, affix := range affixes {
		if position == "prefix" {
			if strings.HasPrefix(name, affix) {
				// matches, so trim it and return
				return name[len(affix):], true
			}
		} else if strings.HasSuffix(name, affix) {
			return name[:len(name)-len(affix)], true
		}
	}

	n.report("missingAffix", fmt.Sprintf("%s name `%s` must have one of the following %ses: %s", n.typeName, n.originalName, position, strings.Join(affixes, ", ")))
	return "", false
}

// validateCustom returns true if the name is valid according to the `regex` option, false otherwise
func (n *nameValidationContext) validateCustom(config *normalizedSelector, name string) bool {
	custom := config.custom
	if custom == nil {
		return true
	}

	result := custom.test(name)
	if custom.match && result {
		return true
	}
	if !custom.match && !result {
		return true
	}

	regexMatch := "match"
	if !custom.match {
		regexMatch = "not match"
	}
	n.report("satisfyCustom", fmt.Sprintf("%s name `%s` must %s the RegExp: /%s/u", n.typeName, n.originalName, regexMatch, custom.source))
	return false
}

// validatePredefinedFormat returns true if the name is valid according to the `format` option, false otherwise
func (n *nameValidationContext) validatePredefinedFormat(config *normalizedSelector, name string, modifiers modifierFlags) bool {
	formats := config.format
	if len(formats) == 0 {
		return true
	}

	if modifiers&modifierRequiresQuotes == 0 {
		for _, format := range formats {
			if format.check(name) {
				return true
			}
		}
	}

	formatNames := make([]string, 0, len(formats))
	for _, format := range formats {
		formatNames = append(formatNames, format.String())
	}
	if n.originalName == name {
		n.report("doesNotMatchFormat", fmt.Sprintf("%s name `%s` must match one of the following formats: %s", n.typeName, n.originalName, strings.Join(formatNames, ", ")))
	} else {
		n.report("doesNotMatchFormatTrimmed", fmt.Sprintf("%s name `%s` trimmed as `%s` must match one of the following formats: %s", n.typeName, n.originalName, name, strings.Join(formatNames, ", ")))
	}
	return false
}

const selectorsAllowedToHaveTypes = selectorVariable |
	selectorParameter |
	selectorClassProperty |
	selectorObjectLiteralProperty |
	selectorTypeProperty |
	selectorParameterProperty |
	selectorClassicAccessor

func isCorrectType(ctx rule.RuleContext, node *ast.Node, config *normalizedSelector, selector selectorFlags) bool {
	if config.types == nil {
		return true
	}

	if selectorsAllowedToHaveTypes&selector == 0 {
		return true
	}

	typeChecker := ctx.TypeChecker
	// remove null and undefined from the type, as we don't care about it here
	t := typeChecker.GetNonNullableType(typeChecker.GetTypeAtLocation(node))

	for _, allowedType := range config.types {
		switch allowedType {
		case typeModifierArray:
			if isAllTypesMatch(t, func(t *checker.Type) bool {
				return checker.Checker_isArrayType(typeChecker, t) || checker.IsTupleType(t)
			}) {
				return true
			}

		case typeModifierFunction:
			if isAllTypesMatch(t, func(t *checker.Type) bool {
				return len(utils.GetCallSignatures(typeChecker, t)) > 0
			}) {
				return true
			}

		case typeModifierBoolean, typeModifierNumber, typeModifierString:
			// this will resolve things like true => boolean, 'a' => string and 1 => number
			typeString := typeChecker.TypeToString(checker.Checker_getWidenedType(typeChecker, checker.Checker_getBaseTypeOfLiteralType(typeChecker, t)))
			var allowedTypeString string
			switch allowedType {
			case typeModifierBoolean:
				allowedTypeString = "boolean"
			case typeModifierNumber:
				allowedTypeString = "number"
			case typeModifierString:
				allowedTypeString = "string"
			}
			if typeString == allowedTypeString {
				return true
			}
		}
	}

	return false
}

// isAllTypesMatch returns true if the type (or all union types) in the given type return true for the callback
func isAllTypesMatch(t *checker.Type, cb func(t *checker.Type) bool) bool {
	if utils.IsUnionType(t) {
		return utils.Every(t.Types(), cb)
	}
	return cb(t)
}

/*
These format functions are taken from `tslint-consistent-codestyle/naming-convention`:
https://github.com/ajafff/tslint-consistent-codestyle/blob/ab156cc8881bcc401236d999f4ce034b59039e81/rules/namingConventionRule.ts#L603-L645

The license for the code can be viewed here:
https://github.com/ajafff/tslint-consistent-codestyle/blob/ab156cc8881bcc401236d999f4ce034b59039e81/LICENSE
*/

/*
Why not regex here? Because it's actually really, really difficult to create a regex to handle
all of the unicode cases, and we have many non-english users that use non-english characters.
https://gist.github.com/mathiasbynens/6334847
*/

func firstRune(name string) rune {
	for _, r := range name {
		return r
	}
	return 0
}

func isPascalCase(name string) bool {
	if name == "" {
		return true
	}
	first := firstRune(name)
	return unicode.ToUpper(first) == first && !strings.Contains(name, "_")
}

func isStrictPascalCase(name string) bool {
	if name == "" {
		return true
	}
	first := firstRune(name)
	return unicode.ToUpper(first) == first && hasStrictCamelHumps(name, true)
}

func isCamelCase(name string) bool {
	if name == "" {
		return true
	}
	first := firstRune(name)
	return unicode.ToLower(first) == first && !strings.Contains(name, "_")
}

func isStrictCamelCase(name string) bool {
	if name == "" {
		return true
	}
	first := firstRune(name)
	return unicode.ToLower(first) == first && hasStrictCamelHumps(name, false)
}

func hasStrictCamelHumps(name string, isUpper bool) bool {
	isUppercaseChar := func(r rune) bool {
		return unicode.ToUpper(r) == r && unicode.ToLower(r) != r
	}

	if strings.HasPrefix(name, "_") {
		return false
	}
	for i, r := range name {
		if i == 0 {
			continue
		}
		if r == '_' {
			return false
		}
		if isUpper == isUppercaseChar(r) {
			if isUpper {
				return false
			}
		} else {
			isUpper = !isUpper
		}
	}
	return true
}

func isSnakeCase(name string) bool {
	return name == "" || (name == strings.ToLower(name) && validateUnderscores(name))
}

func isUpperCase(name string) bool {
	return name == "" || (name == strings.ToUpper(name) && validateUnderscores(name))
}

// validateUnderscores checks for leading, trailing and adjacent underscores
func validateUnderscores(name string) bool {
	if strings.HasPrefix(name, "_") {
		return false
	}
	wasUnderscore := false
	for i, r := range name {
		if i == 0 {
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
