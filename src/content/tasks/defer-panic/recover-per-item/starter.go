package main

import "fmt"

func handle(x int) int {
	return 100 / x
}

// processAll обрабатывает все элементы. Сбой одного
// не должен мешать остальным.
func processAll(items []int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("сбой:", r)
		}
	}()

	for _, x := range items {
		fmt.Println(x, "→", handle(x))
	}
}

func main() {
	processAll([]int{5, 0, 20})
	fmt.Println("конец")
}
