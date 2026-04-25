package utils

import (
	"strconv"
	"strings"
)

func QuoteSingleStringLiteral(value string) string {
	var b strings.Builder
	b.Grow(len(value) + 2)
	b.WriteByte('\'')
	for _, ch := range value {
		switch ch {
		case '\t':
			b.WriteString(`\t`)
		case '\v':
			b.WriteString(`\v`)
		case '\f':
			b.WriteString(`\f`)
		case '\b':
			b.WriteString(`\b`)
		case '\r':
			b.WriteString(`\r`)
		case '\n':
			b.WriteString(`\n`)
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '\u2028', '\u2029', '\u0085':
			writeUnicodeEscape(&b, ch)
		default:
			if ch <= '\u001f' {
				writeUnicodeEscape(&b, ch)
			} else {
				b.WriteRune(ch)
			}
		}
	}
	b.WriteByte('\'')
	return b.String()
}

func writeUnicodeEscape(b *strings.Builder, ch rune) {
	hex := strings.ToUpper(strconv.FormatInt(int64(ch), 16))
	b.WriteString(`\u`)
	for i := len(hex); i < 4; i++ {
		b.WriteByte('0')
	}
	b.WriteString(hex)
}
