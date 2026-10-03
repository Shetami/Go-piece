package main

import (
	"slices"
	"strings"
)

// Anagrams возвращает те candidates, что являются анаграммами subject,
// в их исходном порядке и написании. Регистр не важен; слово не считается
// анаграммой самого себя.
func Anagrams(subject string, candidates []string) []string {
	lower := strings.ToLower(subject)
	key := sortedRunes(lower)
	var out []string
	for _, c := range candidates {
		cl := strings.ToLower(c)
		if cl != lower && sortedRunes(cl) == key {
			out = append(out, c)
		}
	}
	return out
}

// sortedRunes — буквы слова по порядку: у анаграмм он одинаковый.
func sortedRunes(s string) string {
	r := []rune(s)
	slices.Sort(r)
	return string(r)
}
