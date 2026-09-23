package main

import "fmt"

// Split делит список на первый элемент и всё остальное.
func Split(parts []string) (head, tail []string) {
	head = parts[:1]
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
