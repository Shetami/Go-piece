package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu sync.Mutex
	n  int
}

// Inc увеличивает счётчик под блокировкой.
func (c Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func main() {
	var c Counter

	for i := 0; i < 1000; i++ {
		c.Inc()
	}

	fmt.Println("счётчик:", c.Value())
}
