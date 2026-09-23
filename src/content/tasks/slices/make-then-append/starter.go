package main

import "fmt"

// squares возвращает квадраты чисел в том же порядке.
func squares(nums []int) []int {
	out := make([]int, len(nums))
	for _, n := range nums {
		out = append(out, n*n)
	}
	return out
}

func main() {
	fmt.Println(squares([]int{1, 2, 3}))
}
