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
	for _, it := range items {
		inv.mu.Lock()
		defer inv.mu.Unlock()
		inv.stock[it]--
	}
}

func main() {
	inv := &Inventory{stock: map[string]int{"книга": 3, "ручка": 5}}
	inv.Reserve([]string{"книга", "ручка"})
	fmt.Println(inv.stock)
}
