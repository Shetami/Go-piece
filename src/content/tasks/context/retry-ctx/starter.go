package main

import (
	"context"
	"time"
)

// Retry вызывает f, пока она не вернёт nil, но не больше attempts раз.
// Между попытками ждёт delay; после последней попытки не ждёт.
// Если ctx отменён (до попытки или во время ожидания), f больше не
// вызывается, а Retry возвращает ошибку, для которой errors.Is(err, ctx.Err()).
// Если попытки кончились — возвращает последнюю ошибку f.
func Retry(ctx context.Context, attempts int, delay time.Duration, f func(context.Context) error) error {
	// ваш код
	return nil
}
