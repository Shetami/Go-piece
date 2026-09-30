package main

import (
	"context"
	"errors"
	"time"
)

// Hedged выполняет страхующие (hedged) запросы.
//
// Сразу вызывает call(ctx, 0). Если за delay ответа нет — запускает
// call(ctx, 1), ещё через delay — call(ctx, 2), и так до attempts попыток
// (attempts >= 1). Если попытка завершилась ошибкой, следующая запускается
// сразу, не дожидаясь delay.
//
// Возвращает первый успешный результат; контекст остальных попыток при этом
// отменяется. Если упали все — errors.Join ошибок в порядке номеров попыток
// (не в порядке их завершения). Если отменён ctx — сразу ctx.Err().
// Hedged не ждёт отменённые попытки, но и не оставляет горутин, навсегда
// заблокированных на отправке результата.
func Hedged[T any](ctx context.Context, attempts int, delay time.Duration,
	call func(ctx context.Context, i int) (T, error)) (T, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // отменяет проигравшие попытки

	type result struct {
		i   int
		v   T
		err error
	}
	// Буфер на все попытки: опоздавшие отправят и завершатся, даже если
	// их никто уже не читает.
	results := make(chan result, attempts)
	errs := make([]error, attempts)
	launched, finished := 0, 0

	timer := time.NewTimer(delay)
	defer timer.Stop()
	launch := func() {
		i := launched
		launched++
		go func() {
			v, err := call(ctx, i)
			results <- result{i, v, err}
		}()
		// Перезапуск таймера так, чтобы работало при любой семантике таймеров.
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(delay)
	}
	launch()

	var zero T
	for {
		var hedge <-chan time.Time
		if launched < attempts {
			hedge = timer.C
		}
		select {
		case <-hedge:
			launch()
		case r := <-results:
			finished++
			if r.err == nil {
				return r.v, nil
			}
			errs[r.i] = r.err
			if finished == attempts {
				return zero, errors.Join(errs...)
			}
			if launched < attempts {
				launch() // упала — не ждём delay
			}
		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}
}
