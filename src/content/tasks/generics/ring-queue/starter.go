package main

// Queue — очередь FIFO на кольцевом буфере. Растёт по мере надобности.
type Queue[T any] struct {
	// ваши поля
}

func (q *Queue[T]) Push(v T) {
	// ваш код
}

// Pop забирает самый старый элемент. На пустой очереди — нулевое значение и false.
func (q *Queue[T]) Pop() (T, bool) {
	// ваш код
	var zero T
	return zero, false
}

func (q *Queue[T]) Len() int {
	// ваш код
	return 0
}
