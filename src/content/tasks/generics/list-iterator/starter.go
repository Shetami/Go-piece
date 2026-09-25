package main

import "iter"

// List — односвязный список. Нулевое значение — пустой список, готовый к работе.
type List[T any] struct {
	// ваши поля
}

// PushBack добавляет v в конец за O(1).
func (l *List[T]) PushBack(v T) {
	// ваш код
}

// Len возвращает число элементов за O(1).
func (l *List[T]) Len() int {
	// ваш код
	return 0
}

// All возвращает итератор по элементам от начала к концу:
// for v := range l.All() { ... }
func (l *List[T]) All() iter.Seq[T] {
	// ваш код
	return func(yield func(T) bool) {}
}
