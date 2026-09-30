package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrLeaseLost = errors.New("аренда потеряна")

// WithLease возвращает контекст для работы под арендой (лидерство,
// распределённая блокировка). Раз в interval вызывается renew с
// контекстом, производным от результата и ограниченным interval.
//   - renew вернул ошибку (в том числе по своему таймауту) — результат
//     отменяется с причиной, отвечающей errors.Is и на ErrLeaseLost, и на
//     ошибку renew; продления прекращаются;
//   - отмена parent отменяет результат с причиной родителя;
//   - cancel отменяет результат (причина — context.Canceled) и
//     возвращается только после того, как фоновая горутина вышла: после
//     cancel renew больше не выполняется. Повторный cancel безопасен.
//     Вызывать cancel из самого renew нельзя.
func WithLease(parent context.Context, interval time.Duration, renew func(context.Context) error) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(parent)
	done := make(chan struct{})
	go func() {
		defer close(done)
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
			// Продление не должно висеть дольше интервала: зависшее
			// продление — то же, что потерянная аренда.
			rctx, rcancel := context.WithTimeout(ctx, interval)
			err := renew(rctx)
			rcancel()
			if err != nil {
				// Если нас уже отменили, cancel ничего не перезапишет:
				// первая причина остаётся.
				cancel(fmt.Errorf("%w: %w", ErrLeaseLost, err))
				return
			}
		}
	}()
	return ctx, func() {
		cancel(context.Canceled)
		<-done // гарантия: после возврата renew больше не выполняется
	}
}
