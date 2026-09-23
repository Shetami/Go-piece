package main

import (
	"maps"
	"sync"
)

// Counter — счётчики по ключам, безопасные для использования из многих горутин.
type Counter struct {
	mu sync.Mutex
	m  map[string]int
}

func NewCounter() *Counter {
	return &Counter{m: make(map[string]int)}
}

func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key]++
}

func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[key]
}

// Snapshot возвращает копию всех счётчиков. Изменения копии
// не должны влиять на Counter, и наоборот.
func (c *Counter) Snapshot() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Отдать c.m наружу нельзя: вызывающий читал бы её без мьютекса.
	return maps.Clone(c.m)
}
