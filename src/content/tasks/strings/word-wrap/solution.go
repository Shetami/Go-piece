package main

import (
	"strings"
	"unicode/utf8"
)

// Wrap разбивает текст на строки не длиннее width символов (рун).
// Слова разделяются любыми пробельными символами, в результате — одним
// пробелом. Слова не режутся: слово длиннее width идёт отдельной строкой.
func Wrap(text string, width int) []string {
	var lines []string
	var cur strings.Builder
	curLen := 0 // длина текущей строки в рунах: len(cur) считал бы байты
	for _, w := range strings.Fields(text) {
		wl := utf8.RuneCountInString(w)
		if curLen > 0 && curLen+1+wl > width {
			lines = append(lines, cur.String())
			cur.Reset()
			curLen = 0
		}
		if curLen > 0 {
			cur.WriteByte(' ')
			curLen++
		}
		cur.WriteString(w)
		curLen += wl
	}
	if curLen > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}
