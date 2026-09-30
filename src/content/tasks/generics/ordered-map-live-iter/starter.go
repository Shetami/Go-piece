package main

import "iter"

// OrderedMap — мапа, которая помнит порядок вставки ключей. Нулевое
// значение готово к работе; копировать после использования нельзя.
type OrderedMap[K comparable, V any] struct {
	// ваши поля
}

// Set записывает значение. Новый ключ встаёт в конец, существующий
// остаётся на своём месте. O(1).
func (m *OrderedMap[K, V]) Set(k K, v V) {
	// ваш код
}

// Get возвращает значение ключа.
func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

// Delete удаляет ключ; false, если его не было. O(1).
func (m *OrderedMap[K, V]) Delete(k K) bool {
	// ваш код
	return false
}

// Len возвращает число ключей.
func (m *OrderedMap[K, V]) Len() int {
	// ваш код
	return 0
}

// All обходит пары в порядке вставки. Во время обхода можно вызывать
// Set и Delete: удалённые и ещё не выданные ключи не выдаются, новые
// ключи выдаются в конце обхода, каждый живой ключ — ровно один раз.
func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	// ваш код
	return func(yield func(K, V) bool) {}
}
