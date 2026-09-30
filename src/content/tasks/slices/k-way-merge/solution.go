package main

import "container/heap"

// cursor — позиция чтения в одном из списков.
type cursor struct{ list, pos int }

// cursorHeap — min-куча курсоров по текущему элементу; при равенстве выше
// тот, у кого меньше номер списка, — это и даёт стабильность.
type cursorHeap[T any] struct {
	cur   []cursor
	lists [][]T
	cmp   func(a, b T) int
}

func (h *cursorHeap[T]) Len() int { return len(h.cur) }
func (h *cursorHeap[T]) Less(i, j int) bool {
	a, b := h.cur[i], h.cur[j]
	if c := h.cmp(h.lists[a.list][a.pos], h.lists[b.list][b.pos]); c != 0 {
		return c < 0
	}
	return a.list < b.list
}
func (h *cursorHeap[T]) Swap(i, j int) { h.cur[i], h.cur[j] = h.cur[j], h.cur[i] }
func (h *cursorHeap[T]) Push(x any)   { h.cur = append(h.cur, x.(cursor)) }
func (h *cursorHeap[T]) Pop() any {
	x := h.cur[len(h.cur)-1]
	h.cur = h.cur[:len(h.cur)-1]
	return x
}

// MergeK сливает k слайсов, каждый из которых отсортирован по cmp, в один
// отсортированный слайс. Слияние стабильное: равные по cmp элементы идут
// в порядке номеров списков, а внутри списка — в исходном порядке.
// Списки не меняются; nil и пустые списки допустимы.
// Сравнений — O(N log k), где N — всего элементов: используйте container/heap.
func MergeK[T any](lists [][]T, cmp func(a, b T) int) []T {
	h := &cursorHeap[T]{lists: lists, cmp: cmp}
	total := 0
	for i, l := range lists {
		total += len(l)
		if len(l) > 0 { // пустые списки в кучу не кладём
			h.cur = append(h.cur, cursor{i, 0})
		}
	}
	heap.Init(h)
	out := make([]T, 0, total)
	for h.Len() > 0 {
		top := &h.cur[0]
		out = append(out, lists[top.list][top.pos])
		top.pos++
		if top.pos < len(lists[top.list]) {
			heap.Fix(h, 0) // вершина сдвинулась — просеиваем вниз, O(log k)
		} else {
			heap.Pop(h) // список кончился
		}
	}
	return out
}
