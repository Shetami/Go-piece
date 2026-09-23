package main

import (
	"fmt"
	"slices"
)

// combine возвращает все сочетания по k элементов из items.
func combine(items []int, k int) [][]int {
	var res [][]int
	var walk func(start int, path []int)
	walk = func(start int, path []int) {
		if len(path) == k {
			// path делит массив со всеми остальными ветками перебора —
			// в результат кладём снимок, а не сам слайс.
			res = append(res, slices.Clone(path))
			return
		}
		for i := start; i < len(items); i++ {
			walk(i+1, append(path, items[i]))
		}
	}
	walk(0, make([]int, 0, k))
	return res
}

func main() {
	for _, c := range combine([]int{1, 2, 3, 4}, 2) {
		fmt.Println(c)
	}
}
