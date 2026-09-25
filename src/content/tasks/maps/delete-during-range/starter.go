package main

import "fmt"

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}

	iters := 0
	for k := range m {
		iters++
		// На первой же итерации удаляем все ключи, кроме текущего.
		for other := range m {
			if other != k {
				delete(m, other)
			}
		}
	}

	fmt.Println("итераций:", iters)
	fmt.Println("осталось:", len(m))
}
