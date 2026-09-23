package main

import (
	"strings"
	"unicode"
)

// ToSnake переводит camelCase и PascalCase в snake_case.
// Аббревиатуры остаются одним словом: "parseURL" → "parse_url",
// "HTTPServer" → "http_server".
func ToSnake(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i, r := range rs {
		if i > 0 && unicode.IsUpper(r) {
			prev := rs[i-1]
			nextLower := i+1 < len(rs) && unicode.IsLower(rs[i+1])
			// Граница слова: «aB» (конец строчного слова) или «ABc»
			// (последняя заглавная аббревиатуры начинает новое слово).
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower) {
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
