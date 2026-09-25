package main

import (
	"context"
	"fmt"
	"time"
)

// Retry вызывает f, пока она не вернёт nil, но не больше attempts раз.
// Между попытками ждёт delay; после последней попытки не ждёт.
// Если ctx отменён (до попытки или во время ожидания), f больше не
// вызывается, а Retry возвращает ошибку, для которой errors.Is(err, ctx.Err()).
// Если попытки кончились — возвращает последнюю ошибку f.
func Retry(ctx context.Context, attempts int, delay time.Duration, f func(context.Context) error) error {
	var last error
	for i := range attempts {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry: попытка %d: %w", i+1, err)
		}
		if last = f(ctx); last == nil {
			return nil
		}
		if i == attempts-1 {
			break // после последней попытки ждать незачем
		}
		// time.Sleep не прерывается отменой — ждём в select с таймером.
		t := time.NewTimer(delay)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return fmt.Errorf("retry: ожидание после попытки %d: %w", i+1, ctx.Err())
		}
	}
	return last
}
