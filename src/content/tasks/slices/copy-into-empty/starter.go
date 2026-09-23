package main

import "fmt"

// clone возвращает независимую копию слайса.
func clone(src []int) []int {
	dst := make([]int, 0, len(src))
	copy(dst, src)
	return dst
}

func main() {
	orig := []int{1, 2, 3}
	c := clone(orig)
	fmt.Println(c, len(c))
}
