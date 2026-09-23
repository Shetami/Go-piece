package main

import "fmt"

// squares возвращает квадраты чисел в том же порядке.
func squares(nums []int) []int {
	// Длина 0, вместимость len(nums): память выделена заранее,
	// а append пишет с начала, а не после нулей.
	out := make([]int, 0, len(nums))
	for _, n := range nums {
		out = append(out, n*n)
	}
	return out
}

func main() {
	fmt.Println(squares([]int{1, 2, 3}))
}
