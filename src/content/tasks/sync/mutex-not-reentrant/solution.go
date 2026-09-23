package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu   sync.Mutex
	data map[string]int
}

func (c *Cache) Get(k string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[k]
}

func (c *Cache) Incr(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = c.Get(k) + 1
}

func main() {
	c := &Cache{data: map[string]int{"a": 1}}
	fmt.Println("старт")
	c.Incr("a")
	fmt.Println("финиш", c.Get("a"))
}
