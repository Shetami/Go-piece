package main

import (
	"strings"
	"unicode"
)

// WordCount считает, сколько раз встречается каждое слово.
// Слова приводятся к нижнему регистру, апостроф внутри слова
// (don't, they're) — часть слова.
func WordCount(sentence string) map[string]int {
	// Режем по всему, что не буква, не цифра и не апостроф.
	words := strings.FieldsFunc(strings.ToLower(sentence), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\''
	})
	counts := make(map[string]int)
	for _, w := range words {
		// Апостроф по краям — это кавычка: 'large' → large.
		if w = strings.Trim(w, "'"); w != "" {
			counts[w]++
		}
	}
	return counts
}
