package main

import (
	"cmp"
	"fmt"
)

// Max возвращает наибольший элемент непустого слайса.
func Max[T cmp.Ordered](xs []T) T {
	// Начинаем с первого элемента, а не с нулевого значения типа:
	// ноль может оказаться больше всех элементов.
	m := xs[0]
	for _, x := range xs[1:] {
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
