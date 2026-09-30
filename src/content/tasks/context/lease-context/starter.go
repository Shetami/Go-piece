package main

import (
	"context"
	"errors"
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
	// ваш код
	return context.WithCancel(parent)
}
