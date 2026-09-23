package main

import (
	"fmt"
	"slices"
)

// removeEmpty удаляет пустые строки, сохраняя порядок остальных.
func removeEmpty(items []string) []string {
	// DeleteFunc проходит слайс один раз и сдвигает оставшиеся элементы сам —
	// индексы не съезжают под ногами.
	return slices.DeleteFunc(items, func(s string) bool { return s == "" })
}

func main() {
	fmt.Printf("%q\n", removeEmpty([]string{"a", "", "", "b", "", "c"}))
}
