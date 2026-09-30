package main

import "container/list"

// LFU — кэш на capacity записей с вытеснением наименее часто используемой.
//
//   - Частота записи — сколько раз к ней обращались: Put нового ключа даёт 1,
//     каждый Get-попадание и Put существующего ключа прибавляют 1.
//   - При переполнении вытесняется запись с наименьшей частотой, а среди
//     равных — та, к которой дольше всего не обращались.
//   - Вытесненный и вставленный заново ключ начинает с частоты 1.
//   - capacity <= 0 — кэш ничего не хранит.
//   - Get и Put — за O(1).
type LFU[K comparable, V any] struct {
	capacity int
	items    map[K]*list.Element // ключ → узел в списке своей частоты
	buckets  map[int]*list.List  // частота → записи, спереди самые свежие
	minFreq  int
}

type lfuEntry[K comparable, V any] struct {
	key  K
	val  V
	freq int
}

func NewLFU[K comparable, V any](capacity int) *LFU[K, V] {
	return &LFU[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		buckets:  make(map[int]*list.List),
	}
}

func (c *LFU[K, V]) bucket(f int) *list.List {
	l := c.buckets[f]
	if l == nil {
		l = list.New()
		c.buckets[f] = l
	}
	return l
}

// touch переносит запись в список частоты freq+1.
func (c *LFU[K, V]) touch(el *list.Element) {
	e := el.Value.(*lfuEntry[K, V])
	old := c.buckets[e.freq]
	old.Remove(el)
	if old.Len() == 0 {
		delete(c.buckets, e.freq)
		if c.minFreq == e.freq {
			c.minFreq++ // запись ушла на freq+1 — она и есть новый минимум
		}
	}
	e.freq++
	c.items[e.key] = c.bucket(e.freq).PushFront(e)
}

func (c *LFU[K, V]) Get(k K) (V, bool) {
	el, ok := c.items[k]
	if !ok {
		var zero V
		return zero, false
	}
	c.touch(el)
	return c.items[k].Value.(*lfuEntry[K, V]).val, true
}

func (c *LFU[K, V]) Put(k K, v V) {
	if c.capacity <= 0 {
		return
	}
	if el, ok := c.items[k]; ok {
		el.Value.(*lfuEntry[K, V]).val = v
		c.touch(el)
		return
	}
	if len(c.items) >= c.capacity {
		// Среди самых редких — самая давняя: хвост списка minFreq.
		l := c.buckets[c.minFreq]
		victim := l.Back()
		l.Remove(victim)
		if l.Len() == 0 {
			delete(c.buckets, c.minFreq)
		}
		delete(c.items, victim.Value.(*lfuEntry[K, V]).key)
	}
	c.items[k] = c.bucket(1).PushFront(&lfuEntry[K, V]{key: k, val: v, freq: 1})
	c.minFreq = 1 // новая запись всегда самая редкая
}

func (c *LFU[K, V]) Len() int { return len(c.items) }
