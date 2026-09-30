package main

import (
	"context"
	"errors"
)

var ErrClosed = errors.New("queue closed")

// Queue — очередь задач с фиксированным числом воркеров и буфером.
type Queue struct {
	// ваши поля
}

// NewQueue запускает workers воркеров; в буфере помещается size задач.
func NewQueue(workers, size int) *Queue {
	// ваш код
	return &Queue{}
}

// Submit ставит задачу в очередь. Если буфер полон — ждёт места.
// Возвращает ctx.Err(), если ctx отменён раньше; ErrClosed, если Shutdown
// уже начат (в том числе пока Submit ждал места). Безопасен одновременно
// с Shutdown. Принятая задача (nil) обязательно будет выполнена.
func (q *Queue) Submit(ctx context.Context, task func()) error {
	// ваш код
	return nil
}

// Shutdown перестаёт принимать задачи, дожидается выполнения всех уже
// принятых (включая стоящие в буфере) и остановки воркеров.
// Если ctx истёк раньше — возвращает ctx.Err(), воркеры доделывают в фоне.
// Повторные и одновременные вызовы безопасны.
func (q *Queue) Shutdown(ctx context.Context) error {
	// ваш код
	return nil
}
