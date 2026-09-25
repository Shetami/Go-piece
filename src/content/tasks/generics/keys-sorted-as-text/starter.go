package main

import (
	"fmt"
	"sort"
)

// SortedKeys возвращает ключи мапы по возрастанию.
func SortedKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j]) })
	return keys
}

func main() {
	floors := map[int]string{2: "бухгалтерия", 10: "директор", 1: "ресепшн", 9: "склад"}
	for _, f := range SortedKeys(floors) {
		fmt.Println(f, floors[f])
	}
}
