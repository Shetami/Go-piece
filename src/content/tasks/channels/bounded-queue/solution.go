package main

import (
	"context"
	"errors"
)

var ErrFull = errors.New("queue is full")

// Queue — FIFO-очередь на capacity элементов (capacity >= 1), безопасная
// для любого числа горутин. Создаётся через NewQueue.
type Queue[T any] struct {
	// Буферизованный канал и есть очередь: FIFO, ограничение размера
	// и потокобезопасность достаются бесплатно.
	ch chan T
}

func NewQueue[T any](capacity int) *Queue[T] {
	return &Queue[T]{ch: make(chan T, capacity)}
}

// TryPush кладёт v, никогда не блокируясь. Если очередь полна — ErrFull
// (вызывающий, например, отвечает клиенту 503 вместо того, чтобы копить запросы).
func (q *Queue[T]) TryPush(v T) error {
	select {
	case q.ch <- v:
		return nil
	default: // буфер полон — отказ, а не ожидание
		return ErrFull
	}
}

// PushWait кладёт v, дожидаясь места. Если ctx отменён раньше — ctx.Err().
func (q *Queue[T]) PushWait(ctx context.Context, v T) error {
	select {
	case q.ch <- v:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Pop забирает самый старый элемент, дожидаясь его появления.
// Если ctx отменён раньше — нулевое значение и ctx.Err(). Если элемент
// уже лежит в очереди, он отдаётся даже при отменённом ctx.
func (q *Queue[T]) Pop(ctx context.Context) (T, error) {
	// Сначала без ожидания: если готовы обе ветки, select выбрал бы случайно.
	select {
	case v := <-q.ch:
		return v, nil
	default:
	}
	select {
	case v := <-q.ch:
		return v, nil
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// Len — сколько элементов сейчас в очереди.
func (q *Queue[T]) Len() int {
	return len(q.ch)
}
