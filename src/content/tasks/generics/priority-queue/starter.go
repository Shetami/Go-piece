package main

// PQ — очередь с приоритетом на двоичной куче. Pop всегда снимает
// наименьший по less элемент. Push и Pop — O(log n).
// Пакет container/heap не использовать.
type PQ[T any] struct {
	// ваши поля
}

func NewPQ[T any](less func(a, b T) bool) *PQ[T] {
	return &PQ[T]{}
}

func (q *PQ[T]) Len() int {
	// ваш код
	return 0
}

func (q *PQ[T]) Push(v T) {
	// ваш код
}

// Pop снимает и возвращает наименьший элемент. Для пустой очереди —
// нулевое значение и false.
func (q *PQ[T]) Pop() (T, bool) {
	// ваш код
	var zero T
	return zero, false
}
