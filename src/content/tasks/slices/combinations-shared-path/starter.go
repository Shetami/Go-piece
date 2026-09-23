package main

import "fmt"

// combine возвращает все сочетания по k элементов из items.
func combine(items []int, k int) [][]int {
	var res [][]int
	var walk func(start int, path []int)
	walk = func(start int, path []int) {
		if len(path) == k {
			res = append(res, path)
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
