package main

import (
	"context"
	"fmt"
	"runtime/debug"
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
	for n := 1; ; n++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err, pe := runWorker(ctx, worker)
		if pe == nil {
			return err // nil или обычная ошибка — перезапуск не нужен
		}
		onPanic(n, pe)
		if n > maxRestarts {
			return fmt.Errorf("воркер упал %d раз, перезапуски кончились: %w", n, pe)
		}
		if err := sleepCtx(ctx, backoff(n)); err != nil {
			return err
		}
	}
}

// runWorker — один запуск: recover и снятие стека в том же defer.
func runWorker(ctx context.Context, worker func(context.Context) error) (err error, pe *PanicError) {
	defer func() {
		if r := recover(); r != nil {
			pe = &PanicError{Value: r, Stack: debug.Stack()}
		}
	}()
	return worker(ctx), nil
}

// sleepCtx — пауза, которую можно прервать. time.Sleep так не умеет.
func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
