package main

import "fmt"

// Split делит список на первый элемент и всё остальное.
func Split(parts []string) (head, tail []string) {
	// Третье число в срезе ограничивает вместимость: у head она равна длине,
	// поэтому append вынужден завести новый массив и не затрёт parts[1].
	head = parts[:1:1]
	tail = parts[1:]
	return head, tail
}

func main() {
	parts := []string{"go", "is", "fun"}

	head, tail := Split(parts)
	head = append(head, "!")

	fmt.Println("head:", head)
	fmt.Println("tail:", tail)
}
