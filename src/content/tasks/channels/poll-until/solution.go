package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrNotReady — check возвращает её (или обёртку над ней), когда ресурс
// ещё не готов и стоит спросить позже.
var ErrNotReady = errors.New("not ready")

// PollUntil вызывает check сразу, а потом раз в interval, пока не
// дождётся результата:
//   - check вернул nil — PollUntil возвращает его значение и nil;
//   - check вернул ошибку, для которой errors.Is(err, ErrNotReady), —
//     опрос продолжается;
//   - любая другая ошибка — опрос прекращается, ошибка возвращается как есть.
//
// Если ctx кончился раньше, возвращается ошибка, для которой errors.Is
// верно и с ctx.Err(), и с последней ошибкой «не готов» (если check
// успел её вернуть). Если ctx отменён до вызова, check не вызывается.
// Ресурсы (тикер) освобождаются в любом случае.
func PollUntil[T any](ctx context.Context, interval time.Duration, check func(context.Context) (T, error)) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastErr error
	for {
		// Первая попытка — сразу, а не через interval.
		v, err := check(ctx)
		switch {
		case err == nil:
			return v, nil
		case !errors.Is(err, ErrNotReady):
			return zero, err
		}
		lastErr = err

		select {
		case <-ticker.C:
			// Если check дольше interval, тикер не копит пропущенные тики:
			// в его канале буфер на один.
		case <-ctx.Done():
			// Два %w (Go 1.20+): видны и причина остановки, и почему не дождались.
			return zero, fmt.Errorf("poll: %w: %w", ctx.Err(), lastErr)
		}
	}
}
