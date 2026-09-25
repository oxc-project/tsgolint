package naming_convention

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func jsUpper(s string) string { return cases.Upper(language.Und).String(s) }
func jsLower(s string) string { return cases.Lower(language.Und).String(s) }

// JS indexes a string by UTF-16 code unit. An astral character's first code
// unit is a surrogate, and both case transforms leave that surrogate alone.
func jsFirstCase(name string, upper bool) bool {
	r, _ := utf8.DecodeRuneInString(name)
	if r > 0xffff {
		return true
	}
	first := string(r)
	if upper {
		return first == jsUpper(first)
	}
	return first == jsLower(first)
}

func jsUppercaseChar(r rune) bool {
	if r > 0xffff {
		return false
	}
	s := string(r)
	return s == jsUpper(s) && s != jsLower(s)
}

func strictCamelHumps(name string, isUpper bool) bool {
	if strings.HasPrefix(name, "_") {
		return false
	}
	// Start at the second UTF-16 code unit, as the upstream implementation does.
	_, size := utf8.DecodeRuneInString(name)
	if size == 0 {
		return true
	}
	if first, _ := utf8.DecodeRuneInString(name); first > 0xffff {
		// The low surrogate of the first astral rune is processed by the JS loop.
		if isUpper {
			isUpper = false
		}
	}
	for _, r := range name[size:] {
		if r == '_' {
			return false
		}
		if isUpper == jsUppercaseChar(r) {
			if isUpper {
				return false
			}
		} else {
			isUpper = !isUpper
		}
		if r > 0xffff {
			// Its low surrogate is another non-uppercase UTF-16 code unit.
			if isUpper {
				isUpper = false
			}
		}
	}
	return true
}

func validInternalUnderscores(name string) bool {
	if strings.HasPrefix(name, "_") {
		return false
	}
	previous := false
	for _, r := range name {
		if r == '_' {
			if previous {
				return false
			}
			previous = true
		} else {
			previous = false
		}
	}
	return !previous
}

func matchesNamingFormat(name, format string) bool {
	if name == "" {
		return true
	}
	switch format {
	case "camelCase":
		return jsFirstCase(name, false) && !strings.Contains(name, "_")
	case "strictCamelCase":
		return jsFirstCase(name, false) && strictCamelHumps(name, false)
	case "PascalCase":
		return jsFirstCase(name, true) && !strings.Contains(name, "_")
	case "StrictPascalCase":
		return jsFirstCase(name, true) && strictCamelHumps(name, true)
	case "snake_case":
		return name == jsLower(name) && validInternalUnderscores(name)
	case "UPPER_CASE":
		return name == jsUpper(name) && validInternalUnderscores(name)
	}
	return false
}
