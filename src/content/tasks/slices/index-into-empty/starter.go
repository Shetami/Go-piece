package main

import "fmt"

// firstSquares возвращает квадраты чисел от 1 до n.
func firstSquares(n int) []int {
	out := make([]int, 0, n)
	for i := range n {
		out[i] = (i + 1) * (i + 1)
	}
	return out
}

func main() {
	fmt.Println(firstSquares(5))
}
