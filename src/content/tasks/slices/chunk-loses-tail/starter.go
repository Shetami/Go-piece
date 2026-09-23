package main

import "fmt"

// chunk режет слайс на куски по size элементов; последний кусок может быть короче.
func chunk(xs []int, size int) [][]int {
	var out [][]int
	for i := 0; i+size <= len(xs); i += size {
		out = append(out, xs[i:i+size])
	}
	return out
}

func main() {
	fmt.Println(chunk([]int{1, 2, 3, 4, 5}, 2))
	fmt.Println(chunk([]int{1, 2, 3, 4}, 2))
	fmt.Println(chunk([]int{1}, 3))
}
