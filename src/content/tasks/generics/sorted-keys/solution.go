package main

import (
	"cmp"
	"maps"
	"slices"
)

// SortedKeys возвращает ключи мапы по возрастанию.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	// maps.Keys — итератор, slices.Sorted собирает его и сортирует.
	return slices.Sorted(maps.Keys(m))
}
