package main

import (
	"cmp"
	"iter"
)

// Tree — двоичное дерево поиска (без балансировки). Нулевое значение
// готово к работе.
type Tree[K cmp.Ordered, V any] struct {
	// ваши поля
}

// Put записывает значение; существующий ключ перезаписывается.
func (t *Tree[K, V]) Put(key K, val V) {
	// ваш код
}

// Get возвращает значение ключа и true, если ключ есть.
func (t *Tree[K, V]) Get(key K) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

// Delete удаляет ключ и возвращает true, если он был.
func (t *Tree[K, V]) Delete(key K) bool {
	// ваш код
	return false
}

// Len возвращает число ключей.
func (t *Tree[K, V]) Len() int {
	// ваш код
	return 0
}

// All обходит все пары по возрастанию ключа.
func (t *Tree[K, V]) All() iter.Seq2[K, V] {
	// ваш код
	return func(yield func(K, V) bool) {}
}

// Range обходит пары с lo <= key < hi по возрастанию ключа и не
// заходит в поддеревья, где подходящих ключей быть не может.
func (t *Tree[K, V]) Range(lo, hi K) iter.Seq2[K, V] {
	// ваш код
	return func(yield func(K, V) bool) {}
}
