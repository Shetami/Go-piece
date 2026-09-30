package main

import (
	"container/heap"
	"iter"
)

// head — текущий элемент одного источника.
type head[T any] struct {
	val  T
	src  int // номер источника: при равенстве раньше идёт меньший
	next func() (T, bool)
}

type heads[T any] struct {
	items []head[T]
	cmp   func(a, b T) int
}

func (h *heads[T]) Len() int { return len(h.items) }
func (h *heads[T]) Less(i, j int) bool {
	if c := h.cmp(h.items[i].val, h.items[j].val); c != 0 {
		return c < 0
	}
	return h.items[i].src < h.items[j].src // стабильность между источниками
}
func (h *heads[T]) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *heads[T]) Push(x any)   { h.items = append(h.items, x.(head[T])) }
func (h *heads[T]) Pop() any {
	it := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]
	return it
}

// Merge сливает последовательности, каждая из которых отсортирована по
// cmp, в одну отсортированную. Равные элементы идут в порядке номеров
// источников, внутри источника — в исходном порядке. Слияние ленивое:
// из каждого источника берётся не больше одного элемента впрок. Когда
// обход закончен или прерван break, все источники остановлены.
func Merge[T any](cmp func(a, b T) int, seqs ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		h := &heads[T]{cmp: cmp}
		for i, s := range seqs {
			next, stop := iter.Pull(s)
			// stop обязателен: без него горутина-сопрограмма источника
			// и его defer так и не выполнятся после break.
			defer stop()
			if v, ok := next(); ok {
				h.items = append(h.items, head[T]{v, i, next})
			}
		}
		heap.Init(h)
		for h.Len() > 0 {
			top := &h.items[0]
			if !yield(top.val) {
				return
			}
			if v, ok := top.next(); ok {
				top.val = v
				heap.Fix(h, 0) // заменили вершину — погружаем
			} else {
				heap.Pop(h)
			}
		}
	}
}
