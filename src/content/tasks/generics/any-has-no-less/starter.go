package main

import "fmt"

// Max возвращает наибольший из аргументов.
func Max[T any](first T, rest ...T) T {
	m := first
	for _, x := range rest {
		if x > m {
			m = x
		}
	}
	return m
}

func main() {
	fmt.Println(Max(3, 7, 2))
	fmt.Println(Max("груша", "яблоко", "абрикос"))
	fmt.Println(Max(2.5, -1.0))
}
