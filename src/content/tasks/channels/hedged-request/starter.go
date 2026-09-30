package main

import (
	"context"
	"errors"
	"time"
)

var ErrNoFetchers = errors.New("no fetchers")

// Hedge делает «подстрахованный» запрос. Сначала запускается fetchers[0].
// Каждый следующий запускается, когда с запуска предыдущего прошло delay
// и ответа всё нет, — или сразу, как только какая-то попытка упала.
//
// Первый успешный ответ возвращается немедленно, контекст остальных
// попыток отменяется; ждать их Hedge не должен, но и висеть их горутины
// не должны. Если упали все — возвращается errors.Join всех ошибок в
// порядке fetchers. Если отменён ctx — ctx.Err(). Без fetchers —
// ErrNoFetchers.
func Hedge[T any](ctx context.Context, delay time.Duration, fetchers ...func(context.Context) (T, error)) (T, error) {
	// ваш код
	var zero T
	return zero, nil
}
