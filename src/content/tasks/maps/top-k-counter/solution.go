package main

import (
	"cmp"
	"container/heap"
	"slices"
)

// Entry — ключ и его счётчик.
type Entry[K cmp.Ordered] struct {
	Key   K
	Count int
}

// Counter считает события по ключам.
//
//   - Add прибавляет delta (может быть отрицательной). Ключ, чей счётчик
//     стал ≤ 0, удаляется.
//   - Count — текущий счётчик, 0 для отсутствующего ключа.
//   - Len — число ключей с положительным счётчиком.
//   - TopK — до k записей с наибольшими счётчиками, по убыванию счётчика,
//     при равенстве — по возрастанию ключа. k ≤ 0 — пустой результат.
//     Не меняет счётчики. Сложность — O(u log k), где u — число ключей:
//     используйте кучу на k элементов (container/heap), а не сортировку всех.
type Counter[K cmp.Ordered] struct {
	counts map[K]int
}

func NewCounter[K cmp.Ordered]() *Counter[K] {
	return &Counter[K]{counts: make(map[K]int)}
}

func (c *Counter[K]) Add(key K, delta int) {
	n := c.counts[key] + delta
	if n <= 0 {
		delete(c.counts, key) // нулевые и отрицательные в топ не попадают
		return
	}
	c.counts[key] = n
}

func (c *Counter[K]) Count(key K) int { return c.counts[key] }

func (c *Counter[K]) Len() int { return len(c.counts) }

// better сообщает, должна ли a стоять в топе выше b.
func better[K cmp.Ordered](a, b Entry[K]) bool {
	if a.Count != b.Count {
		return a.Count > b.Count
	}
	return a.Key < b.Key
}

// worstHeap — куча, на вершине которой худшая из отобранных записей:
// её и выталкиваем, когда находится кто-то лучше.
type worstHeap[K cmp.Ordered] []Entry[K]

func (h worstHeap[K]) Len() int           { return len(h) }
func (h worstHeap[K]) Less(i, j int) bool { return better(h[j], h[i]) }
func (h worstHeap[K]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *worstHeap[K]) Push(x any)        { *h = append(*h, x.(Entry[K])) }
func (h *worstHeap[K]) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func (c *Counter[K]) TopK(k int) []Entry[K] {
	if k <= 0 {
		return nil
	}
	h := make(worstHeap[K], 0, min(k, len(c.counts)))
	for key, n := range c.counts {
		e := Entry[K]{key, n}
		if h.Len() < k {
			heap.Push(&h, e)
		} else if better(e, h[0]) {
			h[0] = e
			heap.Fix(&h, 0)
		}
	}
	// В куче k лучших, но в порядке кучи: досортировываем k элементов.
	res := []Entry[K](h)
	slices.SortFunc(res, func(a, b Entry[K]) int {
		if better(a, b) {
			return -1
		}
		if better(b, a) {
			return 1
		}
		return 0
	})
	return res
}
