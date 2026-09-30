package main

import (
	"context"
	"errors"
)

var ErrFull = errors.New("queue is full")

// Queue — FIFO-очередь на capacity элементов (capacity >= 1), безопасная
// для любого числа горутин. Создаётся через NewQueue.
type Queue[T any] struct {
	// ваши поля
}

func NewQueue[T any](capacity int) *Queue[T] {
	// ваш код
	return &Queue[T]{}
}

// TryPush кладёт v, никогда не блокируясь. Если очередь полна — ErrFull
// (вызывающий, например, отвечает клиенту 503 вместо того, чтобы копить запросы).
func (q *Queue[T]) TryPush(v T) error {
	// ваш код
	return nil
}

// PushWait кладёт v, дожидаясь места. Если ctx отменён раньше — ctx.Err().
func (q *Queue[T]) PushWait(ctx context.Context, v T) error {
	// ваш код
	return nil
}

// Pop забирает самый старый элемент, дожидаясь его появления.
// Если ctx отменён раньше — нулевое значение и ctx.Err(). Если элемент
// уже лежит в очереди, он отдаётся даже при отменённом ctx.
func (q *Queue[T]) Pop(ctx context.Context) (T, error) {
	// ваш код
	var zero T
	return zero, nil
}

// Len — сколько элементов сейчас в очереди.
func (q *Queue[T]) Len() int {
	// ваш код
	return 0
}
