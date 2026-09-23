package main

import "fmt"

type point struct{ x, y int }

func main() {
	seen := map[point]bool{}
	seen[point{1, 2}] = true
	fmt.Println(seen[point{1, 2}], seen[point{2, 1}])

	grid := map[[2]int]string{{0, 0}: "старт"}
	k := [2]int{0, 0}
	fmt.Println(grid[k])
	k[0] = 5
	fmt.Println(grid[k] == "", len(grid))

	byIface := map[any]int{}
	byIface[1] = 1
	byIface[int64(1)] = 2
	byIface[1.0] = 3
	byIface[1] = 4
	fmt.Println(len(byIface), byIface[1])
}
