package main

import (
	"fmt"
	"slices"
)

// insertAt вставляет v в позицию i и возвращает результат.
func insertAt(s []int, i, v int) []int {
	// tail делил бы массив с s, и append(s[:i], v) затёр бы его первый элемент.
	return slices.Insert(s, i, v)
}

func main() {
	nums := []int{1, 2, 3, 4, 5}
	nums = insertAt(nums, 2, 99)
	fmt.Println(nums)
	nums = insertAt(nums, 0, 0)
	fmt.Println(nums)
	nums = insertAt(nums, len(nums), 100)
	fmt.Println(nums)
}
