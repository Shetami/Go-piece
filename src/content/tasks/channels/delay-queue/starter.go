package main

import (
	"context"
	"time"
)

// DelayQueue — очередь отложенных задач (повторы с задержкой, напоминания,
// истечение TTL). Элемент становится доступен через delay после Put.
// Безопасна для многих производителей и потребителей.
type DelayQueue[T any] struct {
	// ваши поля
}

func NewDelayQueue[T any]() *DelayQueue[T] {
	// ваш код
	return &DelayQueue[T]{}
}

// Put добавляет v, который станет доступен через delay (delay <= 0 — сразу).
func (q *DelayQueue[T]) Put(v T, delay time.Duration) {
	// ваш код
}

// Pop ждёт и возвращает элемент с самым ранним временем готовности; при
// равном времени — тот, что добавлен раньше. Если, пока Pop ждёт, добавлен
// элемент, готовый раньше, Pop должен вернуть его вовремя. Если ctx
// отменён раньше — нулевое значение и ctx.Err().
func (q *DelayQueue[T]) Pop(ctx context.Context) (T, error) {
	// ваш код
	var zero T
	return zero, nil
}

// Len — сколько элементов в очереди (готовых и нет).
func (q *DelayQueue[T]) Len() int {
	// ваш код
	return 0
}
