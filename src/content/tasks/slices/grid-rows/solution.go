package main

import "fmt"

func newGrid(h, w int) [][]rune {
	grid := make([][]rune, h)
	for y := range grid {
		// make для внешнего слайса создал только h пустых строк — каждую выделяем отдельно.
		grid[y] = make([]rune, w)
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
