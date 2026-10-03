package main

import "unicode"

// IsIsogram сообщает, что в фразе ни одна буква не повторяется.
// Регистр не важен, пробелы и дефисы — не буквы и могут повторяться.
func IsIsogram(phrase string) bool {
	seen := make(map[rune]bool)
	for _, r := range phrase {
		if !unicode.IsLetter(r) {
			continue
		}
		r = unicode.ToLower(r)
		if seen[r] {
			return false
		}
		seen[r] = true
	}
	return true
}
