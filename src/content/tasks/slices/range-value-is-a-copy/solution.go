package main

import "fmt"

type Item struct {
	Name  string
	Price int
}

// Discount снижает цену каждой позиции на 10.
func Discount(items []Item) {
	// Обращаемся к элементу по индексу: item в range — копия,
	// и правки в ней до слайса не доходят.
	for i := range items {
		items[i].Price -= 10
	}
}

func main() {
	cart := []Item{{"книга", 100}, {"кофе", 50}}

	Discount(cart)

	fmt.Println(cart)
}
