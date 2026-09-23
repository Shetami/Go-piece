package main

import "fmt"

// identity строит единичную матрицу n×n.
func identity(n int) [][]int {
	// Каждой строке — свой массив: Repeat скопировал бы заголовок
	// одной и той же строки n раз.
	grid := make([][]int, n)
	for i := range grid {
		grid[i] = make([]int, n)
	}
	for i := range n {
		grid[i][i] = 1
	}
	return grid
}

func main() {
	for _, row := range identity(3) {
		fmt.Println(row)
	}
}
