package main

import (
	"container/heap"
	"iter"
	"slices"
)

type ranked[T any] struct {
	v   T
	idx int // номер во входе: при равенстве раньше встреченный лучше
}

// worstFirst — куча, где в корне худший из отобранных.
type worstFirst[T any] struct {
	items []ranked[T]
	less  func(a, b T) bool
}

// worse: a хуже b — меньше по less, а при равенстве встречен позже.
func (h *worstFirst[T]) worse(a, b ranked[T]) bool {
	if h.less(a.v, b.v) {
		return true
	}
	if h.less(b.v, a.v) {
		return false
	}
	return a.idx > b.idx
}

func (h *worstFirst[T]) Len() int           { return len(h.items) }
func (h *worstFirst[T]) Less(i, j int) bool { return h.worse(h.items[i], h.items[j]) }
func (h *worstFirst[T]) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *worstFirst[T]) Push(x any)         { h.items = append(h.items, x.(ranked[T])) }
func (h *worstFirst[T]) Pop() any {
	x := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]
	return x
}

// TopK возвращает k наибольших по less элементов seq — от большего к
// меньшему; из равных раньше идёт встреченный раньше. Если элементов
// меньше k — все. k <= 0 — nil. Дополнительная память — O(k): seq может
// быть сколь угодно длинной, и целиком она не сохраняется.
func TopK[T any](seq iter.Seq[T], k int, less func(a, b T) bool) []T {
	if k <= 0 {
		return nil
	}
	h := &worstFirst[T]{items: make([]ranked[T], 0, k), less: less}
	i := 0
	for v := range seq {
		cur := ranked[T]{v, i}
		i++
		if len(h.items) < k {
			h.items = append(h.items, cur)
			if len(h.items) == k {
				heap.Init(h) // кучу строим один раз, когда набрали k
			}
			continue
		}
		// Новый элемент встречен позже всех, поэтому при равенстве он
		// хуже корня — заменяем только строго большим.
		if less(h.items[0].v, v) {
			h.items[0] = cur
			heap.Fix(h, 0)
		}
	}
	// Итоговый порядок: лучшие первыми.
	slices.SortFunc(h.items, func(a, b ranked[T]) int {
		switch {
		case h.worse(b, a):
			return -1
		case h.worse(a, b):
			return 1
		}
		return 0
	})
	out := make([]T, len(h.items))
	for j, r := range h.items {
		out[j] = r.v
	}
	return out
}
