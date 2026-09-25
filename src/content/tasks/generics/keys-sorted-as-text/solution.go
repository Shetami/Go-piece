package main

import (
	"cmp"
	"fmt"
	"slices"
)

// SortedKeys возвращает ключи мапы по возрастанию.
// cmp.Ordered даёт настоящее сравнение < для чисел и строк — без перевода в текст.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func main() {
	floors := map[int]string{2: "бухгалтерия", 10: "директор", 1: "ресепшн", 9: "склад"}
	for _, f := range SortedKeys(floors) {
		fmt.Println(f, floors[f])
	}
}
