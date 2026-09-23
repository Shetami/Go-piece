package main

import "container/list"

// LRU — кэш фиксированной ёмкости. При переполнении вытесняется запись,
// к которой дольше всего не обращались (и Get, и Put считаются обращением).
type LRU struct {
	capacity int
	order    *list.List               // спереди — самые свежие
	items    map[string]*list.Element // ключ → узел списка, за O(1)
}

type entry struct {
	key   string
	value int
}

func NewLRU(capacity int) *LRU {
	return &LRU{capacity: capacity, order: list.New(), items: make(map[string]*list.Element)}
}

func (c *LRU) Get(key string) (int, bool) {
	el, ok := c.items[key]
	if !ok {
		return 0, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*entry).value, true
}

func (c *LRU) Put(key string, value int) {
	if el, ok := c.items[key]; ok {
		el.Value.(*entry).value = value
		c.order.MoveToFront(el)
		return
	}
	c.items[key] = c.order.PushFront(&entry{key, value})
	if c.order.Len() > c.capacity {
		// Самый старый — в хвосте. Ключ храним в узле как раз для этого:
		// чтобы удалить запись и из мапы.
		oldest := c.order.Back()
		c.order.Remove(oldest)
		delete(c.items, oldest.Value.(*entry).key)
	}
}

func (c *LRU) Len() int { return c.order.Len() }
