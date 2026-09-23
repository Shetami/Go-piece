package main

import (
	"context"
	"time"
)

// RunWithTimeout запускает f с контекстом, у которого таймаут d, и
// возвращает её ошибку. Если f не уложилась, RunWithTimeout возвращает
// context.DeadlineExceeded сразу по истечении срока, не дожидаясь f —
// даже если f контекст игнорирует.
func RunWithTimeout(ctx context.Context, d time.Duration, f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	// Буфер на одно значение: если мы уйдём по таймауту, f всё равно
	// сможет отдать результат и завершиться, а не повиснет на отправке.
	done := make(chan error, 1)
	go func() { done <- f(ctx) }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
