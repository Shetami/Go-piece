package main

import "cmp"

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
	// ваши поля
}

func NewCounter[K cmp.Ordered]() *Counter[K] {
	return &Counter[K]{}
}

func (c *Counter[K]) Add(key K, delta int) {
	// ваш код
}

func (c *Counter[K]) Count(key K) int {
	// ваш код
	return 0
}

func (c *Counter[K]) Len() int {
	// ваш код
	return 0
}

func (c *Counter[K]) TopK(k int) []Entry[K] {
	// ваш код
	return nil
}
