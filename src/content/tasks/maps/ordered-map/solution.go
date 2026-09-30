package main

import "iter"

// OrderedMap — мапа, которая помнит порядок вставки ключей.
//
//   - Set нового ключа ставит его в конец; Set существующего меняет только
//     значение, место ключа не меняется.
//   - Delete — за O(1), возвращает, был ли ключ. Ключ, удалённый и
//     добавленный заново, встаёт в конец.
//   - All обходит пары в порядке вставки и корректно останавливается по break.
//     Во время обхода можно вызывать Delete для любых ключей (в том числе
//     текущего) и Set для существующих: удалённые и ещё не пройденные ключи
//     не выдаются, значения выдаются актуальные. Ключи, добавленные во время
//     обхода, в этот обход не попадают.
type OrderedMap[K comparable, V any] struct {
	items      map[K]*omNode[K, V]
	head, tail *omNode[K, V]
	seq        uint64 // номер следующей вставки
}

type omNode[K comparable, V any] struct {
	key        K
	val        V
	prev, next *omNode[K, V]
	seq        uint64
	removed    bool
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{items: make(map[K]*omNode[K, V])}
}

func (m *OrderedMap[K, V]) Set(k K, v V) {
	if n, ok := m.items[k]; ok {
		n.val = v // место не меняем
		return
	}
	m.seq++
	n := &omNode[K, V]{key: k, val: v, prev: m.tail, seq: m.seq}
	if m.tail != nil {
		m.tail.next = n
	} else {
		m.head = n
	}
	m.tail = n
	m.items[k] = n
}

func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
	if n, ok := m.items[k]; ok {
		return n.val, true
	}
	var zero V
	return zero, false
}

func (m *OrderedMap[K, V]) Delete(k K) bool {
	n, ok := m.items[k]
	if !ok {
		return false
	}
	delete(m.items, k)
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		m.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		m.tail = n.prev
	}
	// Указатели удалённого узла НЕ обнуляем: итератор, который стоит на нём,
	// пойдёт по n.next дальше. Флаг говорит «не выдавать».
	n.removed = true
	return true
}

func (m *OrderedMap[K, V]) Len() int { return len(m.items) }

func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		// Всё, что вставлено после начала обхода, имеет seq больше этого.
		limit := m.seq
		for n := m.head; n != nil && n.seq <= limit; n = n.next {
			if n.removed {
				continue
			}
			if !yield(n.key, n.val) {
				return
			}
		}
	}
}
