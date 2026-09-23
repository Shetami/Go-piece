package main

// LRU — кэш фиксированной ёмкости. При переполнении вытесняется запись,
// к которой дольше всего не обращались (и Get, и Put считаются обращением).
type LRU struct {
	// ваши поля
}

func NewLRU(capacity int) *LRU {
	return &LRU{}
}

func (c *LRU) Get(key string) (int, bool) {
	// ваш код
	return 0, false
}

func (c *LRU) Put(key string, value int) {
	// ваш код
}

func (c *LRU) Len() int {
	// ваш код
	return 0
}
