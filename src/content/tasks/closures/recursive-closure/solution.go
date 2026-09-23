package main

import "fmt"

func main() {
	// Сначала объявляем переменную, потом присваиваем ей функцию —
	// тогда внутри тела имя fib уже существует.
	var fib func(n int) int
	fib = func(n int) int {
		if n < 2 {
			return n
		}
		return fib(n-1) + fib(n-2)
	}
	fmt.Println(fib(10))
}
