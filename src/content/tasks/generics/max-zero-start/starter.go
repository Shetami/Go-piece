package main

import (
	"cmp"
	"fmt"
)

// Max возвращает наибольший элемент непустого слайса.
func Max[T cmp.Ordered](xs []T) T {
	var m T
	for _, x := range xs {
		if x > m {
			m = x
		}
	}
	return m
}

func main() {
	fmt.Println(Max([]int{3, 7, 2}))
	fmt.Println(Max([]int{-5, -2, -9}))
	fmt.Println(Max([]float64{-0.5, -1.5}))
}
