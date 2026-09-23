package main

import "fmt"

// Index возвращает позицию target в xs или -1.
func Index[T any](xs []T, target T) int {
	for i, x := range xs {
		if x == target {
			return i
		}
	}
	return -1
}

func main() {
	fmt.Println(Index([]string{"a", "b"}, "b"))
	fmt.Println(Index([]int{1, 2, 3}, 5))
}
