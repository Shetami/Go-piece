package main

import (
	"fmt"
	"slices"
	"sort"
)

// median возвращает медиану, не трогая исходные данные.
func median(xs []int) int {
	// Присваивание слайса копирует только заголовок — сортировать надо копию данных.
	sorted := slices.Clone(xs)
	sort.Ints(sorted)
	return sorted[len(sorted)/2]
}

func main() {
	temps := []int{30, 10, 20}
	fmt.Println("медиана:", median(temps))
	fmt.Println("исходные:", temps)
}
