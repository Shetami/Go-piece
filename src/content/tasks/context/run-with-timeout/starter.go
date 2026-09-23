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
	// ваш код
	return f(ctx)
}
