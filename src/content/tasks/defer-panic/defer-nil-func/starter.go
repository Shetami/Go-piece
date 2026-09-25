package main

import "fmt"

func process(items []string) {
	var onDone func()
	defer onDone()

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
