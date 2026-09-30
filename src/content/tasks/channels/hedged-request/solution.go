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
	var zero T
	if len(fetchers) == 0 {
		return zero, ErrNoFetchers
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // отменит проигравших, когда мы вернёмся

	type result struct {
		i   int
		v   T
		err error
	}
	// Буфер на все попытки: опоздавшие отправят результат и завершатся,
	// даже когда читать их уже некому.
	results := make(chan result, len(fetchers))
	errs := make([]error, len(fetchers))
	launched, failed := 0, 0

	var timer <-chan time.Time
	launch := func() {
		i := launched
		launched++
		go func() {
			v, err := fetchers[i](ctx)
			results <- result{i, v, err}
		}()
		timer = nil
		if launched < len(fetchers) {
			timer = time.After(delay) // отсчёт — от последнего запуска
		}
	}
	launch()

	for {
		select {
		case r := <-results:
			if r.err == nil {
				return r.v, nil
			}
			errs[r.i] = r.err
			failed++
			if failed == len(fetchers) {
				return zero, errors.Join(errs...)
			}
			if launched < len(fetchers) {
				launch() // упала — подстраховка нужна сейчас, а не через delay
			}
		case <-timer:
			launch()
		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}
}
