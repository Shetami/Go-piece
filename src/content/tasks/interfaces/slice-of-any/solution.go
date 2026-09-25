package main

import "fmt"

// Параметр типа вместо []any: подходит слайс любого типа, без копирования.
func printAll[T any](items []T) {
	for _, it := range items {
		fmt.Println(it)
	}
}

func main() {
	names := []string{"аня", "боря"}
	ids := []int{1, 2}

	printAll(names)
	printAll(ids)
}
