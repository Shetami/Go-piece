package main

import "fmt"

// insertAt вставляет v в позицию i и возвращает результат.
func insertAt(s []int, i, v int) []int {
	tail := s[i:]
	s = append(s[:i], v)
	return append(s, tail...)
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
