package main

import (
	"fmt"
	"strconv"
	"sync"
)

type Cache struct {
	mu sync.Mutex
	m  map[string]int
}

func (c *Cache) Set(k string, v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[k] = v
}

func (c *Cache) Get(k string) int {
	return c.m[k]
}

func main() {
	c := &Cache{m: map[string]int{}}

	var wg sync.WaitGroup
	for i := range 1000 {
		key := strconv.Itoa(i % 10)
		wg.Go(func() { c.Set(key, i) })
		wg.Go(func() { _ = c.Get(key) })
	}
	wg.Wait()

	fmt.Println("ключей:", len(c.m))
}
