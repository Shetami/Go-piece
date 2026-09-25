package main

import "fmt"

type Number interface {
	~int | ~int8 | ~float64
}

func Sum[T Number](xs ...T) T {
	var s T
	for _, x := range xs {
		s += x
	}
	return s
}

func Avg[T Number](xs ...T) T {
	return Sum(xs...) / T(len(xs))
}

func main() {
	fmt.Println(Avg(1, 2))
	fmt.Println(Avg(1.0, 2))
	fmt.Println(Avg[int8](100, 100))
	fmt.Println(Sum(0.1, 0.2))
}
