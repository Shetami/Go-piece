package main

import "fmt"

var open int

type conn struct{ id int }

func dial(id int) *conn {
	open++
	fmt.Println("открыто соединений:", open)
	return &conn{id: id}
}

func (c *conn) Close() { open-- }

// process обрабатывает задачи по одной, держа не больше одного соединения.
func process(ids []int) {
	for _, id := range ids {
		handle(id)
	}
}

// handle — отдельная функция ради defer: он срабатывает при выходе из неё,
// то есть в конце каждой итерации, а не в конце всего цикла.
func handle(id int) {
	c := dial(id)
	defer c.Close()
	_ = c.id // работа с соединением
}

func main() {
	process([]int{1, 2, 3, 4})
	fmt.Println("после:", open)
}
