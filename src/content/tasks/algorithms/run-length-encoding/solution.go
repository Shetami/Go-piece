package main

import (
	"strconv"
	"strings"
	"unicode"
)

// Encode сжимает строку: серия из n > 1 одинаковых символов
// превращается в число n и символ, одиночный символ остаётся как есть.
func Encode(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i := 0; i < len(runes); {
		// j — первый символ за концом серии.
		j := i + 1
		for j < len(runes) && runes[j] == runes[i] {
			j++
		}
		if n := j - i; n > 1 {
			b.WriteString(strconv.Itoa(n))
		}
		b.WriteRune(runes[i])
		i = j
	}
	return b.String()
}

// Decode разворачивает то, что сделал Encode.
func Decode(s string) string {
	var b strings.Builder
	n := 0
	for _, r := range s {
		if unicode.IsDigit(r) {
			n = n*10 + int(r-'0')
			continue
		}
		// Числа перед символом не было — значит, символ один.
		b.WriteString(strings.Repeat(string(r), max(n, 1)))
		n = 0
	}
	return b.String()
}
