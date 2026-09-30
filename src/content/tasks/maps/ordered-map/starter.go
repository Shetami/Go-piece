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
	// ваши поля
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{}
}

func (m *OrderedMap[K, V]) Set(k K, v V) {
	// ваш код
}

func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

func (m *OrderedMap[K, V]) Delete(k K) bool {
	// ваш код
	return false
}

func (m *OrderedMap[K, V]) Len() int {
	// ваш код
	return 0
}

func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		// ваш код
	}
}
