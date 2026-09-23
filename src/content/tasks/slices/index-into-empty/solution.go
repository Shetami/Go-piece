package main

import "fmt"

// firstSquares возвращает квадраты чисел от 1 до n.
func firstSquares(n int) []int {
	// Вместимость n, длина 0: писать можно только через append.
	out := make([]int, 0, n)
	for i := range n {
		out = append(out, (i+1)*(i+1))
	}
	return out
}

func main() {
	fmt.Println(firstSquares(5))
}
