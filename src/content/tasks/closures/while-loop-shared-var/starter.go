package main

import "fmt"

func main() {
	var tasks []func()

	i := 0
	for i < 3 {
		tasks = append(tasks, func() {
			fmt.Println("задача", i)
		})
		i++
	}

	for _, t := range tasks {
		t()
	}
}
