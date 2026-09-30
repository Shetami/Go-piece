package main

import "errors"

// ErrClosed — очередь закрыта.
var ErrClosed = errors.New("queue closed")

// Queue — ограниченная блокирующая FIFO-очередь между производителями и
// потребителями (как буферизованный канал, но с Len и честным закрытием).
type Queue[T any] struct {
	// ваши поля
}

// NewQueue создаёт очередь на capacity элементов (capacity >= 1).
func NewQueue[T any](capacity int) *Queue[T] {
	// ваш код
	return &Queue[T]{}
}

// Put кладёт элемент, ожидая свободного места. Если очередь закрыта —
// до вызова или пока Put ждал места, — возвращает ErrClosed.
func (q *Queue[T]) Put(v T) error {
	// ваш код
	return nil
}

// Take забирает самый старый элемент, ожидая, пока он появится.
// После Close отдаёт оставшиеся элементы, а когда их нет — (zero, false).
func (q *Queue[T]) Take() (T, bool) {
	// ваш код
	var zero T
	return zero, false
}

// Close закрывает очередь и будит всех ждущих. Повторный Close ничего не делает.
func (q *Queue[T]) Close() {
	// ваш код
}

// Len — сколько элементов сейчас в очереди.
func (q *Queue[T]) Len() int {
	// ваш код
	return 0
}
