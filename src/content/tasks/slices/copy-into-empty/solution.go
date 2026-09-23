package main

import "fmt"

// clone возвращает независимую копию слайса.
func clone(src []int) []int {
	// copy пишет только в пределах длины dst, поэтому длина нужна сразу.
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

func main() {
	orig := []int{1, 2, 3}
	c := clone(orig)
	fmt.Println(c, len(c))
}
