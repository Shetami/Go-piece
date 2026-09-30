package main

import (
	"context"
	"fmt"
	"time"
)

// PanicError — паника воркера, превращённая в ошибку.
type PanicError struct {
	Value any
	Stack []byte
}

func (e *PanicError) Error() string { return fmt.Sprintf("паника воркера: %v", e.Value) }

// Supervise запускает worker(ctx) и перезапускает его после паник.
//   - worker вернул nil — Supervise возвращает nil.
//   - worker вернул ошибку — вернуть её, без перезапуска.
//   - worker запаниковал — onPanic(n, pe) (n — номер паники с 1, в pe —
//     значение и стек runtime/debug.Stack), затем пауза backoff(n) и
//     перезапуск. Паника номер maxRestarts+1 не перезапускается:
//     Supervise возвращает ошибку, из которой errors.As достаёт *PanicError
//     этой последней паники.
//   - Пауза прерывается отменой ctx: тогда (и если ctx отменён перед
//     запуском) Supervise сразу возвращает ctx.Err().
func Supervise(ctx context.Context, maxRestarts int, backoff func(n int) time.Duration,
	onPanic func(n int, pe *PanicError), worker func(ctx context.Context) error) error {
	// ваш код
	return nil
}
