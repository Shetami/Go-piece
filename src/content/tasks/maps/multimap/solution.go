package main

import "slices"

// MultiMap хранит для каждого ключа набор значений без повторов,
// в порядке первого добавления.
//
//   - Add добавляет v к ключу k; false, если такая пара уже есть.
//   - Remove убирает пару; false, если её не было. Ключ, у которого не
//     осталось значений, исчезает совсем.
//   - RemoveAll убирает ключ целиком и возвращает его значения.
//   - Get возвращает значения ключа — копию: изменения результата
//     не должны влиять на мапу. Для отсутствующего ключа — пустой результат.
//   - Len — число ключей, Size — число пар.
type MultiMap[K, V comparable] struct {
	buckets map[K]*bucket[V]
	size    int
}

// bucket — значения одного ключа: слайс держит порядок,
// мапа отвечает «есть ли такое значение» за O(1).
type bucket[V comparable] struct {
	order []V
	set   map[V]struct{}
}

func NewMultiMap[K, V comparable]() *MultiMap[K, V] {
	return &MultiMap[K, V]{buckets: make(map[K]*bucket[V])}
}

func (m *MultiMap[K, V]) Add(k K, v V) bool {
	b := m.buckets[k]
	if b == nil {
		b = &bucket[V]{set: make(map[V]struct{})}
		m.buckets[k] = b
	}
	if _, dup := b.set[v]; dup {
		return false
	}
	b.set[v] = struct{}{}
	b.order = append(b.order, v)
	m.size++
	return true
}

func (m *MultiMap[K, V]) Remove(k K, v V) bool {
	b := m.buckets[k]
	if b == nil {
		return false
	}
	if _, ok := b.set[v]; !ok {
		return false
	}
	delete(b.set, v)
	b.order = slices.DeleteFunc(b.order, func(x V) bool { return x == v })
	m.size--
	if len(b.set) == 0 {
		delete(m.buckets, k) // пустой ключ не должен считаться живым
	}
	return true
}

func (m *MultiMap[K, V]) RemoveAll(k K) []V {
	b := m.buckets[k]
	if b == nil {
		return nil
	}
	delete(m.buckets, k)
	m.size -= len(b.order)
	return b.order // бакет больше никому не принадлежит — копировать не нужно
}

func (m *MultiMap[K, V]) Get(k K) []V {
	b := m.buckets[k]
	if b == nil {
		return nil
	}
	return slices.Clone(b.order) // наружу — копия, не внутренний слайс
}

func (m *MultiMap[K, V]) Has(k K) bool {
	_, ok := m.buckets[k]
	return ok
}

func (m *MultiMap[K, V]) Len() int  { return len(m.buckets) }
func (m *MultiMap[K, V]) Size() int { return m.size }
