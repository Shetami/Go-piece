package main

import "fmt"

// positive возвращает положительные числа из xs.
// Исходный слайс остаётся без изменений.
func positive(xs []int) []int {
	out := xs[:0]
	for _, x := range xs {
		if x > 0 {
			out = append(out, x)
		}
	}
	return out
}

func main() {
	readings := []int{5, -1, 3, -7, 8}

	pos := positive(readings)
	fmt.Println("положительные:", pos)
	fmt.Println("все показания:", readings)
}
