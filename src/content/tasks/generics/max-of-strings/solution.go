package main

import (
	"cmp"
	"fmt"
	"math"
)

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func main() {
	fmt.Println(Max(10, 9), Max("10", "9"), Max(2.5, 2))
	fmt.Println(max(3, 7, 1), min("b", "a", "c"))
	fmt.Println(Max(math.NaN(), 1), max(math.NaN(), 1))
}
