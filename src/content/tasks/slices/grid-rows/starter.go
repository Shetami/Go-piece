package main

import "fmt"

func newGrid(h, w int) [][]rune {
	grid := make([][]rune, h)
	for y := range grid {
		for x := 0; x < w; x++ {
			grid[y][x] = '.'
		}
	}
	return grid
}

func main() {
	g := newGrid(3, 4)
	g[1][2] = '#'
	for _, row := range g {
		fmt.Println(string(row))
	}
}
