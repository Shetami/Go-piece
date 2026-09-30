package main

import "iter"

type omEntry[K comparable, V any] struct {
	key        K
	val        V
	prev, next *omEntry[K, V]
	deleted    bool // удалённый узел держит старые ссылки — по ним уходит итератор
}

// OrderedMap — мапа, которая помнит порядок вставки ключей. Нулевое
// значение готово к работе; копировать после использования нельзя.
type OrderedMap[K comparable, V any] struct {
	idx  map[K]*omEntry[K, V]
	root omEntry[K, V] // страж перед первым элементом, никогда не удаляется
	tail *omEntry[K, V]
}

func (m *OrderedMap[K, V]) last() *omEntry[K, V] {
	if m.tail == nil {
		return &m.root
	}
	return m.tail
}

// Set записывает значение. Новый ключ встаёт в конец, существующий
// остаётся на своём месте. O(1).
func (m *OrderedMap[K, V]) Set(k K, v V) {
	if e, ok := m.idx[k]; ok {
		e.val = v
		return
	}
	if m.idx == nil {
		m.idx = make(map[K]*omEntry[K, V])
	}
	t := m.last()
	e := &omEntry[K, V]{key: k, val: v, prev: t}
	t.next = e
	m.tail = e
	m.idx[k] = e
}

// Get возвращает значение ключа.
func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
	if e, ok := m.idx[k]; ok {
		return e.val, true
	}
	var zero V
	return zero, false
}

// Delete удаляет ключ; false, если его не было. O(1).
func (m *OrderedMap[K, V]) Delete(k K) bool {
	e, ok := m.idx[k]
	if !ok {
		return false
	}
	delete(m.idx, k)
	e.deleted = true
	// Вырезаем из списка, но e.prev и e.next не обнуляем: на e может
	// стоять идущий сейчас обход.
	e.prev.next = e.next
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		m.tail = e.prev
		if m.tail == &m.root {
			m.tail = nil
		}
	}
	return true
}

// Len возвращает число ключей.
func (m *OrderedMap[K, V]) Len() int { return len(m.idx) }

// All обходит пары в порядке вставки. Во время обхода можно вызывать
// Set и Delete: удалённые и ещё не выданные ключи не выдаются, новые
// ключи выдаются в конце обхода, каждый живой ключ — ровно один раз.
func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for nd := m.root.next; nd != nil; {
			if !yield(nd.key, nd.val) {
				return
			}
			// Текущий узел могли удалить внутри yield. Его next устарел
			// (например, он был хвостом, а потом дописали новые), поэтому
			// отступаем к ближайшему живому предку и идём от него.
			for nd.deleted {
				nd = nd.prev
			}
			nd = nd.next
		}
	}
}
