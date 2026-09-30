package main

import (
	"container/heap"
	"sync"
	"time"
)

// TTLCache — потокобезопасный кэш, у каждой записи свой срок жизни.
// Время берётся только из now — так кэш можно тестировать без sleep.
//
//   - Set кладёт значение со сроком ttl; ttl <= 0 — бессрочно. Повторный Set
//     того же ключа заменяет и значение, и срок.
//   - Запись жива, пока now() строго раньше момента истечения. Get истёкшей
//     записи — промах, запись при этом удаляется.
//   - Len — число живых записей на текущий момент.
//   - Purge удаляет все истёкшие записи и возвращает их число. Он не должен
//     обходить все записи: храните сроки в куче (container/heap).
type TTLCache[K comparable, V any] struct {
	mu    sync.Mutex // не RWMutex: Get тоже пишет (удаляет истёкшее)
	now   func() time.Time
	items map[K]*ttlItem[V]
	exp   expHeap[K]
	gen   uint64
}

type ttlItem[V any] struct {
	val     V
	expires time.Time // нулевое — бессрочно
	gen     uint64    // поколение: какая запись в куче актуальна
}

type expEntry[K comparable] struct {
	key     K
	expires time.Time
	gen     uint64
}

type expHeap[K comparable] []expEntry[K]

func (h expHeap[K]) Len() int           { return len(h) }
func (h expHeap[K]) Less(i, j int) bool { return h[i].expires.Before(h[j].expires) }
func (h expHeap[K]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *expHeap[K]) Push(x any)        { *h = append(*h, x.(expEntry[K])) }
func (h *expHeap[K]) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func NewTTLCache[K comparable, V any](now func() time.Time) *TTLCache[K, V] {
	return &TTLCache[K, V]{now: now, items: make(map[K]*ttlItem[V])}
}

func (it *ttlItem[V]) alive(now time.Time) bool {
	return it.expires.IsZero() || now.Before(it.expires)
}

func (c *TTLCache[K, V]) Set(k K, v V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gen++
	it := &ttlItem[V]{val: v, gen: c.gen}
	if ttl > 0 {
		it.expires = c.now().Add(ttl)
		// Старую запись ключа в куче не ищем: она устареет по gen.
		heap.Push(&c.exp, expEntry[K]{k, it.expires, c.gen})
	}
	c.items[k] = it
}

func (c *TTLCache[K, V]) Get(k K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[k]
	if !ok {
		var zero V
		return zero, false
	}
	if !it.alive(c.now()) {
		delete(c.items, k)
		var zero V
		return zero, false
	}
	return it.val, true
}

func (c *TTLCache[K, V]) purge() int {
	now := c.now()
	n := 0
	for len(c.exp) > 0 && !now.Before(c.exp[0].expires) {
		e := heap.Pop(&c.exp).(expEntry[K])
		it, ok := c.items[e.key]
		// Запись в куче могла устареть: ключ перезаписан (другое поколение)
		// или уже удалён через Get. Удаляем только актуальную.
		if ok && it.gen == e.gen {
			delete(c.items, e.key)
			n++
		}
	}
	return n
}

func (c *TTLCache[K, V]) Purge() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.purge()
}

func (c *TTLCache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.purge()
	return len(c.items)
}
