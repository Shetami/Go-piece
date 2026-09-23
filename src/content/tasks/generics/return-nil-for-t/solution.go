package main

import "fmt"

// First возвращает первый элемент и true, или «ничего» и false для пустого слайса.
func First[T any](xs []T) (T, bool) {
	if len(xs) == 0 {
		// Нулевое значение T: 0 для чисел, "" для строк, nil для указателей.
		var zero T
		return zero, false
	}
	return xs[0], true
}

func main() {
	fmt.Println(First([]int{}))
	fmt.Println(First([]string{"го"}))
}
