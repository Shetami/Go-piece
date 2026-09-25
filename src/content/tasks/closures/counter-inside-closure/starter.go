package main

import "fmt"

// newIDGen возвращает генератор идентификаторов: order-1, order-2, ...
func newIDGen(prefix string) func() string {
	return func() string {
		n := 0
		n++
		return fmt.Sprintf("%s-%d", prefix, n)
	}
}

func main() {
	next := newIDGen("order")
	fmt.Println(next(), next(), next())
}
