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
		c := dial(id)
		defer c.Close()
		_ = c.id // работа с соединением
	}
}

func main() {
	process([]int{1, 2, 3, 4})
	fmt.Println("после:", open)
}
