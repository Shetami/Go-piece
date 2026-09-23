package main

import "fmt"

// chunk режет слайс на куски по size элементов; последний кусок может быть короче.
func chunk(xs []int, size int) [][]int {
	var out [][]int
	// Идём, пока есть хоть один элемент, а конец куска обрезаем по длине.
	for i := 0; i < len(xs); i += size {
		end := min(i+size, len(xs))
		out = append(out, xs[i:end:end])
	}
	return out
}

func main() {
	fmt.Println(chunk([]int{1, 2, 3, 4, 5}, 2))
	fmt.Println(chunk([]int{1, 2, 3, 4}, 2))
	fmt.Println(chunk([]int{1}, 3))
}
