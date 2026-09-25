package main

import (
	"slices"
	"strings"
)

// GroupAnagrams группирует слова, составленные из одних и тех же букв.
// Регистр не важен. Слова внутри группы — в порядке появления во входе,
// группы — в порядке появления своего первого слова.
func GroupAnagrams(words []string) [][]string {
	var groups [][]string
	// Ключ — отсортированные руны слова: у анаграмм он одинаковый.
	// Значение — номер группы, чтобы сохранить порядок появления.
	index := map[string]int{}
	for _, w := range words {
		r := []rune(strings.ToLower(w))
		slices.Sort(r)
		key := string(r)
		i, ok := index[key]
		if !ok {
			i = len(groups)
			index[key] = i
			groups = append(groups, nil)
		}
		groups[i] = append(groups[i], w)
	}
	return groups
}
