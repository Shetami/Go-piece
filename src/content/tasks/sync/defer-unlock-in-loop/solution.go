package main

import (
	"fmt"
	"sync"
)

type Inventory struct {
	mu    sync.Mutex
	stock map[string]int
}

// Reserve списывает со склада по одной штуке каждой позиции.
func (inv *Inventory) Reserve(items []string) {
	// Одна блокировка на весь заказ: заодно списание получается атомарным —
	// никто не увидит заказ списанным наполовину.
	inv.mu.Lock()
	defer inv.mu.Unlock()
	for _, it := range items {
		inv.stock[it]--
	}
}

func main() {
	inv := &Inventory{stock: map[string]int{"книга": 3, "ручка": 5}}
	inv.Reserve([]string{"книга", "ручка"})
	fmt.Println(inv.stock)
}
