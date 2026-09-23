package main

import "fmt"

type Integer interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int
}

// Sum складывает числа. Результат — int64, чтобы влезла любая сумма.
func Sum[T Integer](xs []T) int64 {
	// Копим сразу в int64: сумма int8 легко выходит за пределы int8.
	var total int64
	for _, x := range xs {
		total += int64(x)
	}
	return total
}

func main() {
	fmt.Println(Sum([]int{100, 100, 100}))
	fmt.Println(Sum([]int8{100, 100, 100}))
}
