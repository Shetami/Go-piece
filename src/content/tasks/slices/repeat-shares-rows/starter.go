package main

import (
	"fmt"
	"slices"
)

// identity строит единичную матрицу n×n.
func identity(n int) [][]int {
	grid := slices.Repeat([][]int{make([]int, n)}, n)
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
