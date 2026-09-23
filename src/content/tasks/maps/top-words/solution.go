package main

import (
	"cmp"
	"maps"
	"slices"
	"strings"
	"unicode"
)

// TopWords возвращает k самых частых слов текста.
// Регистр не важен, словом считается подряд идущие буквы и цифры.
// При равной частоте слова идут по алфавиту. Слова — в нижнем регистре.
func TopWords(text string, k int) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	freq := make(map[string]int)
	for _, w := range words {
		freq[w]++
	}

	// У мапы нет порядка: выносим ключи и сортируем по частоте, затем по алфавиту.
	keys := slices.Collect(maps.Keys(freq))
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Or(cmp.Compare(freq[b], freq[a]), cmp.Compare(a, b))
	})
	return keys[:min(k, len(keys))]
}
