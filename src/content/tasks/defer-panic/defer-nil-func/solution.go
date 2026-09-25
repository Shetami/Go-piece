package main

import "fmt"

func process(items []string) {
	var onDone func()
	// Отложенная функция вычисляется в момент defer. Оборачиваем в замыкание,
	// чтобы onDone прочиталась при выходе, когда её уже присвоили.
	defer func() { onDone() }()

	start := len(items)
	onDone = func() {
		fmt.Println("обработано:", start)
	}

	for _, it := range items {
		fmt.Println("элемент", it)
	}
}

func main() {
	process([]string{"a", "b"})
	fmt.Println("готово")
}
