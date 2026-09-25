package main

import (
	"fmt"
	"sync"
)

type Stock struct {
	mu    sync.Mutex
	items map[string]int
}

func (s *Stock) Take(name string, n int) error {
	s.mu.Lock()
	// Unlock на любом пути выхода — в том числе на раннем return с ошибкой.
	defer s.mu.Unlock()
	if s.items[name] < n {
		return fmt.Errorf("на складе мало: %s", name)
	}
	s.items[name] -= n
	return nil
}

func main() {
	s := &Stock{items: map[string]int{"чай": 10, "кофе": 3}}
	fmt.Println(s.Take("чай", 5))
	fmt.Println(s.Take("чай", 100))
	fmt.Println(s.Take("кофе", 1))
}
