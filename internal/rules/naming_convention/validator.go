package naming_convention

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dlclark/regexp2/v2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/jsnum"
	"github.com/microsoft/typescript-go/shim/stringutil"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

var namingSelectorBits = map[string]int{
	"variable": 1 << 0, "function": 1 << 1, "parameter": 1 << 2,
	"parameterProperty": 1 << 3, "classicAccessor": 1 << 4, "enumMember": 1 << 5,
	"classMethod": 1 << 6, "objectLiteralMethod": 1 << 7, "typeMethod": 1 << 8,
	"classProperty": 1 << 9, "objectLiteralProperty": 1 << 10, "typeProperty": 1 << 11,
	"autoAccessor": 1 << 12, "class": 1 << 13, "interface": 1 << 14,
	"typeAlias": 1 << 15, "enum": 1 << 16, "typeParameter": 1 << 17, "import": 1 << 18,
}
var namingMetaBits = map[string]int{
	"default":      -1,
	"variableLike": 7,
	"memberLike":   ((1 << 13) - 1) &^ 7,
	"typeLike":     ((1 << 18) - 1) &^ ((1 << 13) - 1),
	"method":       1<<6 | 1<<7 | 1<<8,
	"property":     1<<9 | 1<<10 | 1<<11,
	"accessor":     1<<4 | 1<<12,
}
var namingModifierBits = map[string]int{
	"const": 1 << 0, "readonly": 1 << 1, "static": 1 << 2, "public": 1 << 3,
	"protected": 1 << 4, "private": 1 << 5, "#private": 1 << 6, "abstract": 1 << 7,
	"destructured": 1 << 8, "global": 1 << 9, "exported": 1 << 10,
	"unused": 1 << 11, "requiresQuotes": 1 << 12, "override": 1 << 13,
	"async": 1 << 14, "default": 1 << 15, "namespace": 1 << 16,
}
var namingTypeBits = map[string]int{
	"boolean": 1 << 17, "string": 1 << 18, "number": 1 << 19,
	"function": 1 << 20, "array": 1 << 21,
}

const namingTypeSelectors = 1<<0 | 1<<2 | 1<<9 | 1<<10 | 1<<11 | 1<<3 | 1<<4

type namingMatch struct {
	regex   *regexp2.Regexp
	pattern string
	match   bool
}
type namingConfig struct {
	selector           int
	weight             int
	format             []string
	leadingUnderscore  string
	trailingUnderscore string
	prefix             []string
	suffix             []string
	modifiers          []string
	types              []string
	filter             *namingMatch
	custom             *namingMatch
}
type namingValidator struct {
	ctx             rule.RuleContext
	configs         map[string][]namingConfig
	neededModifiers map[string]bool
}

func namingRegex(pattern string, match bool) *namingMatch {
	// The upstream always compiles with the Unicode flag. regexp2's ECMAScript
	// mode supplies JS character-class and backreference behavior.
	// Frontend option validation cannot bridge regexp2's different Unicode
	// property support, so valid ECMAScript properties must be expanded here.
	compiled, err := regexp2.Compile(namingRegexPattern(pattern), regexp2.ECMAScript|regexp2.Unicode)
	if err != nil {
		panic(fmt.Sprintf("naming-convention: invalid regular expression %q: %v", pattern, err))
	}
	return &namingMatch{regex: compiled, pattern: pattern, match: match}
}

// Adapt regexp2's character matching to ECMAScript /u semantics. Unicode
// properties use the pinned ECMAScript tables, dot excludes all four JS line
// terminators, and word boundaries use ECMAScript's ASCII word set.
func namingRegexPattern(pattern string) string {
	var out strings.Builder
	inClass := false
	for i := 0; i < len(pattern); {
		if pattern[i] == '\\' && i+3 < len(pattern) && (pattern[i+1] == 'p' || pattern[i+1] == 'P') && pattern[i+2] == '{' {
			end := strings.IndexByte(pattern[i+3:], '}')
			if end >= 0 {
				end += i + 3
				property := pattern[i+3 : end]
				if expanded, found := namingExpandUnicodeProperty(property, pattern[i+1] == 'P', inClass); found {
					out.WriteString(expanded)
					i = end + 1
					continue
				}
				out.WriteString(pattern[i : i+3])
				out.WriteString(property)
				out.WriteByte('}')
				i = end + 1
				continue
			}
		}
		if !inClass && pattern[i] == '\\' && i+1 < len(pattern) && (pattern[i+1] == 'b' || pattern[i+1] == 'B') {
			// ECMAScript \w is ASCII-only without the ignore-case flag, so its
			// word boundaries must not use regexp2's Unicode-aware \b semantics.
			if pattern[i+1] == 'b' {
				out.WriteString(`(?:(?<![A-Za-z0-9_])(?=[A-Za-z0-9_])|(?<=[A-Za-z0-9_])(?![A-Za-z0-9_]))`)
			} else {
				out.WriteString(`(?:(?<![A-Za-z0-9_])(?![A-Za-z0-9_])|(?<=[A-Za-z0-9_])(?=[A-Za-z0-9_]))`)
			}
			i += 2
			continue
		}
		if pattern[i] == '\\' && i+1 < len(pattern) {
			out.WriteString(pattern[i : i+2])
			i += 2
			continue
		}
		if pattern[i] == '.' && !inClass {
			out.WriteString(`[^\n\r\u2028\u2029]`)
			i++
			continue
		}
		if pattern[i] == '[' && !inClass {
			inClass = true
		} else if pattern[i] == ']' && inClass {
			inClass = false
		}
		out.WriteByte(pattern[i])
		i++
	}
	return out.String()
}

func namingExpandUnicodeProperty(property string, negated, inClass bool) (string, bool) {
	group, value, hasValue := strings.Cut(property, "=")
	if hasValue {
		if canonical, ok := namingPropertyNameAliases[group]; ok {
			group = canonical
		}
		aliases, ok := namingPropertyValueAliases[group]
		if !ok {
			return "", false
		}
		canonical, ok := aliases[value]
		if !ok {
			return "", false
		}
		value = canonical
	} else {
		if canonical, ok := namingPropertyNameAliases[property]; ok {
			property = canonical
		}
		// Lone property names denote either binary properties or general
		// categories; Script and Script_Extensions always require a value.
		if _, ok := namingPropertyRangesByName["Binary_Property="+property]; ok {
			group, value = "Binary_Property", property
		} else if canonical, ok := namingPropertyValueAliases["General_Category"][property]; ok {
			group, value = "General_Category", canonical
		} else {
			return "", false
		}
	}
	ranges, ok := namingPropertyRangesByName[group+"="+value]
	if !ok {
		return "", false
	}
	content := namingPropertyClassContent(ranges, negated)
	if inClass {
		return content, true
	}
	return "[" + content + "]", true
}

func namingPropertyClassContent(ranges []uint32, negated bool) string {
	var out strings.Builder
	writeRange := func(start, end uint32) {
		out.WriteString(`\u{`)
		out.WriteString(strconv.FormatUint(uint64(start), 16))
		out.WriteByte('}')
		if end-start > 1 {
			out.WriteByte('-')
			out.WriteString(`\u{`)
			out.WriteString(strconv.FormatUint(uint64(end-1), 16))
			out.WriteByte('}')
		}
	}
	if negated {
		cursor := uint32(0)
		for i := 0; i < len(ranges); i += 2 {
			if cursor < ranges[i] {
				writeRange(cursor, ranges[i])
			}
			cursor = ranges[i+1]
		}
		if cursor < 0x110000 {
			writeRange(cursor, 0x110000)
		}
	} else {
		for i := 0; i < len(ranges); i += 2 {
			writeRange(ranges[i], ranges[i+1])
		}
	}
	return out.String()
}

func newNamingValidator(ctx rule.RuleContext, options any) *namingValidator {
	parsed := utils.UnmarshalOptions[NamingConventionOptions](options, "naming-convention")
	if len(parsed) == 0 {
		parsed = NamingConventionOptions{
			{Selector: NamingSelector{"default"}, Format: []string{"camelCase"}, LeadingUnderscore: "allow", TrailingUnderscore: "allow"},
			{Selector: NamingSelector{"import"}, Format: []string{"camelCase", "PascalCase"}},
			{Selector: NamingSelector{"variable"}, Format: []string{"camelCase", "UPPER_CASE"}, LeadingUnderscore: "allow", TrailingUnderscore: "allow"},
			{Selector: NamingSelector{"typeLike"}, Format: []string{"PascalCase"}},
		}
	}
	v := &namingValidator{ctx: ctx, configs: make(map[string][]namingConfig, len(namingSelectorBits)), neededModifiers: make(map[string]bool)}
	for _, option := range parsed {
		weight := 0
		for _, mod := range option.Modifiers {
			weight |= namingModifierBits[mod]
			v.neededModifiers[mod] = true
		}
		for _, typ := range option.Types {
			weight |= namingTypeBits[typ]
		}
		if option.Filter != nil {
			weight |= 1 << 30
		}
		base := namingConfig{weight: weight, format: option.Format, leadingUnderscore: option.LeadingUnderscore, trailingUnderscore: option.TrailingUnderscore, prefix: option.Prefix, suffix: option.Suffix, modifiers: option.Modifiers, types: option.Types}
		if option.Filter != nil {
			base.filter = namingRegex(option.Filter.Regex, option.Filter.Match)
		}
		if option.Custom != nil {
			base.custom = namingRegex(option.Custom.Regex, option.Custom.Match)
		}
		for _, selector := range option.Selector {
			bits, ok := namingSelectorBits[selector]
			if !ok {
				bits, ok = namingMetaBits[selector]
			}
			if !ok {
				continue
			}
			c := base
			c.selector = bits
			for individual, mask := range namingSelectorBits {
				if bits == -1 || bits&mask != 0 {
					v.configs[individual] = append(v.configs[individual], c)
				}
			}
		}
	}
	for individual := range v.configs {
		configs := v.configs[individual]
		sort.SliceStable(configs, func(i, j int) bool {
			a, b := configs[i], configs[j]
			if a.selector == b.selector {
				return a.weight > b.weight
			}
			aMeta, bMeta := isNamingMeta(a.selector), isNamingMeta(b.selector)
			if aMeta != bMeta {
				return !aMeta
			}
			aSpecial, bSpecial := isNamingMethodProperty(a.selector), isNamingMethodProperty(b.selector)
			if aSpecial != bSpecial {
				return aSpecial
			}
			return a.selector > b.selector
		})
		v.configs[individual] = configs
	}
	return v
}
func isNamingMeta(bits int) bool {
	for _, v := range namingMetaBits {
		if bits == v {
			return true
		}
	}
	return false
}
func isNamingMethodProperty(bits int) bool {
	return bits == namingMetaBits["method"] || bits == namingMetaBits["property"]
}
func (v *namingValidator) needsModifier(name string) bool { return v.neededModifiers[name] }

func namingNodeName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		return jsnum.FromString(strings.ReplaceAll(node.Text(), "_", "")).String()
	case ast.KindBigIntLiteral:
		return jsnum.ParseValidBigInt(node.Text()).String()
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(node.Text(), "#")
	case ast.KindRegularExpressionLiteral:
		return namingRegexpLiteralString(node.Text())
	case ast.KindIdentifier, ast.KindStringLiteral:
		return node.Text()
	case ast.KindTrueKeyword:
		return "true"
	case ast.KindFalseKeyword:
		return "false"
	case ast.KindNullKeyword:
		return "null"
	case ast.KindNoSubstitutionTemplateLiteral:
		return "undefined"
	}
	if name, ok := ast.TryGetTextOfPropertyName(node); ok {
		return name
	}
	// The upstream converts non-Identifier keys through `${node.value}`.
	// Template and expression nodes have no value, so JavaScript renders them
	// as the string "undefined".
	return "undefined"
}

// RegExp#toString emits flags in this canonical order, regardless of the
// order used in the source literal.
func namingRegexpLiteralString(text string) string {
	if len(text) < 2 || text[0] != '/' {
		return text
	}
	end := strings.LastIndexByte(text, '/')
	if end <= 0 {
		return text
	}
	flags := text[end+1:]
	var ordered strings.Builder
	for _, flag := range "dgimsuvy" {
		if strings.ContainsRune(flags, flag) {
			ordered.WriteRune(flag)
		}
	}
	return text[:end+1] + ordered.String()
}
func namingDisplaySelector(selector string) string {
	var out strings.Builder
	for i, r := range selector {
		if unicode.IsUpper(r) && i > 0 {
			out.WriteByte(' ')
		}
		if i == 0 {
			r = unicode.ToUpper(r)
		}
		out.WriteRune(r)
	}
	return out.String()
}
func namingRegexString(pattern string) string {
	if pattern == "" {
		return "/(?:)/u"
	}
	var out strings.Builder
	out.WriteByte('/')
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '/':
			backslashes := 0
			for j := i - 1; j >= 0 && pattern[j] == '\\'; j-- {
				backslashes++
			}
			if backslashes%2 == 0 {
				out.WriteByte('\\')
			}
			out.WriteByte('/')
		case '\n':
			out.WriteString("\\n")
		case '\r':
			out.WriteString("\\r")
		case 0xe2:
			if i+2 < len(pattern) && pattern[i+1] == 0x80 && (pattern[i+2] == 0xa8 || pattern[i+2] == 0xa9) {
				if pattern[i+2] == 0xa8 {
					out.WriteString("\\u2028")
				} else {
					out.WriteString("\\u2029")
				}
				i += 2
			} else {
				out.WriteByte(pattern[i])
			}
		default:
			out.WriteByte(pattern[i])
		}
	}
	out.WriteString("/u")
	return out.String()
}
func (m *namingMatch) test(s string) bool {
	var (
		ok  bool
		err error
	)
	if utf8.ValidString(s) {
		ok, err = m.regex.MatchString(s)
	} else {
		// TypeScript strings are UTF-16. typescript-go preserves lone surrogates
		// as WTF-8, which regexp2's string path decodes as RuneError. Decode the
		// JavaScript string explicitly so Unicode property escapes and surrogate
		// escapes see the original code units.
		runes := make([]rune, 0, len(s))
		for i := 0; i < len(s); {
			r, size := stringutil.DecodeJSStringRune(s[i:])
			runes = append(runes, r)
			i += size
		}
		ok, err = m.regex.MatchRunes(runes)
	}
	if err != nil {
		return false
	}
	return ok
}
func (v *namingValidator) report(node *ast.Node, id, desc string) {
	message := rule.RuleMessage{Id: id, Description: desc}
	if parent := node.Parent; parent != nil && (parent.Kind == ast.KindVariableDeclaration || parent.Kind == ast.KindParameter) && parent.Name() == node {
		// ESTree puts a rest parameter's annotation on RestElement, while a
		// normal parameter's annotation belongs to its Identifier.
		if parent.Kind == ast.KindParameter && parent.AsParameterDeclaration().DotDotDotToken != nil {
			v.ctx.ReportNode(node, message)
			return
		}
		end := node.End()
		if annotation := parent.Type(); annotation != nil && annotation.End() > end {
			end = annotation.End()
		} else if optional := parent.QuestionToken(); optional != nil && optional.End() > end {
			end = optional.End()
		}
		if end != node.End() {
			v.ctx.ReportRange(utils.TrimNodeTextRange(v.ctx.SourceFile, node).WithEnd(end), message)
			return
		}
	}
	v.ctx.ReportNode(node, message)
}
func (v *namingValidator) check(node *ast.Node, selector string, modifiers map[string]bool) {
	name := namingNodeName(node)
	configs := v.configs[selector]
	typeName := namingDisplaySelector(selector)
	for _, config := range configs {
		if config.filter != nil && config.filter.test(name) != config.filter.match {
			continue
		}
		allModifiers := true
		for _, modifier := range config.modifiers {
			if !modifiers[modifier] {
				allModifiers = false
				break
			}
		}
		if !allModifiers {
			continue
		}
		if !v.correctType(node, selector, config.types) {
			continue
		}
		processed := name
		if next, ok := v.underscore(node, typeName, name, processed, "leading", config.leadingUnderscore); ok {
			processed = next
		} else {
			return
		}
		if next, ok := v.underscore(node, typeName, name, processed, "trailing", config.trailingUnderscore); ok {
			processed = next
		} else {
			return
		}
		if next, ok := v.affix(node, typeName, name, processed, "prefix", config.prefix); ok {
			processed = next
		} else {
			return
		}
		if next, ok := v.affix(node, typeName, name, processed, "suffix", config.suffix); ok {
			processed = next
		} else {
			return
		}
		if custom := config.custom; custom != nil && custom.test(processed) != custom.match {
			verb := "not match"
			if custom.match {
				verb = "match"
			}
			v.report(node, "satisfyCustom", fmt.Sprintf("%s name `%s` must %s the RegExp: %s", typeName, name, verb, namingRegexString(custom.pattern)))
			return
		}
		if len(config.format) > 0 {
			valid := false
			if !modifiers["requiresQuotes"] {
				for _, format := range config.format {
					if matchesNamingFormat(processed, format) {
						valid = true
						break
					}
				}
			}
			if !valid {
				formats := strings.Join(config.format, ", ")
				if processed == name {
					v.report(node, "doesNotMatchFormat", fmt.Sprintf("%s name `%s` must match one of the following formats: %s", typeName, name, formats))
				} else {
					v.report(node, "doesNotMatchFormatTrimmed", fmt.Sprintf("%s name `%s` trimmed as `%s` must match one of the following formats: %s", typeName, name, processed, formats))
				}
				return
			}
		}
		return
	}
}
func (v *namingValidator) underscore(node *ast.Node, typeName, original, name, position, option string) (string, bool) {
	if option == "" {
		return name, true
	}
	hasSingle := strings.HasPrefix(name, "_")
	hasDouble := strings.HasPrefix(name, "__")
	trimSingle := func() string { return strings.TrimPrefix(name, "_") }
	trimDouble := func() string { return strings.TrimPrefix(name, "__") }
	if position == "trailing" {
		hasSingle = strings.HasSuffix(name, "_")
		hasDouble = strings.HasSuffix(name, "__")
		trimSingle = func() string { return strings.TrimSuffix(name, "_") }
		trimDouble = func() string { return strings.TrimSuffix(name, "__") }
	}
	switch option {
	case "allow":
		if hasSingle {
			return trimSingle(), true
		}
	case "allowDouble":
		if hasDouble {
			return trimDouble(), true
		}
	case "allowSingleOrDouble":
		if hasDouble {
			return trimDouble(), true
		}
		if hasSingle {
			return trimSingle(), true
		}
	case "forbid":
		if hasSingle {
			v.report(node, "unexpectedUnderscore", fmt.Sprintf("%s name `%s` must not have a %s underscore.", typeName, original, position))
			return "", false
		}
	case "require":
		if !hasSingle {
			v.report(node, "missingUnderscore", fmt.Sprintf("%s name `%s` must have one %s underscore(s).", typeName, original, position))
			return "", false
		}
		return trimSingle(), true
	case "requireDouble":
		if !hasDouble {
			v.report(node, "missingUnderscore", fmt.Sprintf("%s name `%s` must have two %s underscore(s).", typeName, original, position))
			return "", false
		}
		return trimDouble(), true
	}
	return name, true
}
func (v *namingValidator) affix(node *ast.Node, typeName, original, name, position string, affixes []string) (string, bool) {
	if len(affixes) == 0 {
		return name, true
	}
	for _, affix := range affixes {
		if position == "prefix" && strings.HasPrefix(name, affix) {
			return strings.TrimPrefix(name, affix), true
		}
		if position == "suffix" && strings.HasSuffix(name, affix) {
			return strings.TrimSuffix(name, affix), true
		}
	}
	v.report(node, "missingAffix", fmt.Sprintf("%s name `%s` must have one of the following %ses: %s", typeName, original, position, strings.Join(affixes, ", ")))
	return "", false
}
func (v *namingValidator) correctType(node *ast.Node, selector string, allowed []string) bool {
	if allowed == nil || namingSelectorBits[selector]&namingTypeSelectors == 0 {
		return true
	}
	t := v.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return false
	}
	t = checker.Checker_GetNonNullableType(v.ctx.TypeChecker, t)
	if t == nil {
		return false
	}
	for _, kind := range allowed {
		switch kind {
		case "array":
			if namingAllUnionTypes(t, func(part *checker.Type) bool {
				return checker.Checker_isArrayType(v.ctx.TypeChecker, part) || checker.IsTupleType(part)
			}) {
				return true
			}
		case "function":
			if namingAllUnionTypes(t, func(part *checker.Type) bool {
				return len(checker.Checker_getSignaturesOfType(v.ctx.TypeChecker, part, checker.SignatureKindCall)) > 0
			}) {
				return true
			}
		case "boolean", "number", "string":
			base := checker.Checker_getBaseTypeOfLiteralType(v.ctx.TypeChecker, t)
			if base == nil {
				continue
			}
			widened := checker.Checker_getWidenedType(v.ctx.TypeChecker, base)
			if widened != nil && v.ctx.TypeChecker.TypeToString(widened) == kind {
				return true
			}
		}
	}
	return false
}
func namingAllUnionTypes(t *checker.Type, predicate func(*checker.Type) bool) bool {
	if t.IsUnion() {
		for _, part := range t.Types() {
			if !predicate(part) {
				return false
			}
		}
		return true
	}
	return predicate(t)
}
