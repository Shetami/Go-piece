package main

import "iter"

type node[T any] struct {
	val  T
	next *node[T]
}

// List — односвязный список. Нулевое значение — пустой список, готовый к работе.
type List[T any] struct {
	head, tail *node[T] // хвост нужен для PushBack за O(1)
	n          int
}

// PushBack добавляет v в конец за O(1).
func (l *List[T]) PushBack(v T) {
	nd := &node[T]{val: v}
	if l.tail == nil {
		l.head = nd
	} else {
		l.tail.next = nd
	}
	l.tail = nd
	l.n++
}

// Len возвращает число элементов за O(1).
func (l *List[T]) Len() int { return l.n }

// All возвращает итератор по элементам от начала к концу:
// for v := range l.All() { ... }
func (l *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for nd := l.head; nd != nil; nd = nd.next {
			// false от yield — цикл сделал break: дальше звать yield нельзя.
			if !yield(nd.val) {
				return
			}
		}
	}
}
