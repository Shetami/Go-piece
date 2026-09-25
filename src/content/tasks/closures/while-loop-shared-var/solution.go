package main

import "fmt"

func main() {
	var tasks []func()

	i := 0
	for i < 3 {
		// i объявлена вне цикла — она одна на все итерации.
		// Копируем значение в переменную, которая своя у каждой итерации.
		n := i
		tasks = append(tasks, func() {
			fmt.Println("задача", n)
		})
		i++
	}

	for _, t := range tasks {
		t()
	}
}
