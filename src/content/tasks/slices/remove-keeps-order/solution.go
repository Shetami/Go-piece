package main

import (
	"fmt"
	"slices"
)

// remove удаляет из очереди элемент с индексом i.
// Порядок остальных элементов должен сохраниться.
func remove(q []string, i int) []string {
	// Сдвигаем хвост на место удалённого, а не переносим последний элемент.
	return slices.Delete(q, i, i+1)
}

func main() {
	queue := []string{"a", "b", "c", "d", "e"}
	queue = remove(queue, 1)
	fmt.Println(queue)
	queue = remove(queue, 0)
	fmt.Println(queue)
}
