package main

import "fmt"

// remove удаляет из очереди элемент с индексом i.
// Порядок остальных элементов должен сохраниться.
func remove(q []string, i int) []string {
	q[i] = q[len(q)-1]
	return q[:len(q)-1]
}

func main() {
	queue := []string{"a", "b", "c", "d", "e"}
	queue = remove(queue, 1)
	fmt.Println(queue)
	queue = remove(queue, 0)
	fmt.Println(queue)
}
