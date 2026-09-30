package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Policy — параметры повторов.
type Policy struct {
	Attempts int           // сколько всего попыток, >= 1
	Base     time.Duration // пауза после первой неудачи, дальше удваивается
	Max      time.Duration // потолок паузы
	// Sleep ждёт d или отмены ctx и возвращает ctx.Err() при отмене.
	// Тесты подменяют его, чтобы не ждать по-настоящему.
	Sleep func(ctx context.Context, d time.Duration) error
}

// Retry вызывает f, пока она не вернёт nil, по таким правилам.
//
// Классификация ошибки f:
//   - временная — если первое звено цепочки, у которого есть метод
//     Temporary() bool, возвращает true. Её повторяем;
//   - любая другая — постоянная: сразу вернуть её как есть.
//
// Паузы: перед повтором номер k (k = 1, 2, …) ждать min(Base·2^(k-1), Max),
// но не меньше RetryAfter(), если в цепочке есть звено с методом
// RetryAfter() time.Duration (сервер сам сказал, когда приходить, — это
// важнее Max). После последней попытки не ждать.
//
// Контекст: если ctx отменён до первой попытки — вернуть ctx.Err(), не
// вызывая f. Если отмена пришла во время паузы или обнаружена перед
// очередной попыткой — вернуть ошибку, для которой errors.Is находит и
// ctx.Err(), и последнюю ошибку f.
//
// Попытки кончились — вернуть ошибку, которая оборачивает последнюю ошибку f
// и упоминает число попыток.
func Retry(ctx context.Context, p Policy, f func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sleep := p.Sleep
	if sleep == nil {
		sleep = sleepCtx
	}
	delay := p.Base
	var last error
	for attempt := 1; ; attempt++ {
		last = f(ctx)
		if last == nil {
			return nil
		}
		if !isTemporary(last) {
			return last
		}
		if attempt >= p.Attempts {
			return fmt.Errorf("после %d попыток: %w", attempt, last)
		}

		wait := min(delay, p.Max)
		var ra interface{ RetryAfter() time.Duration }
		if errors.As(last, &ra) {
			wait = max(wait, ra.RetryAfter())
		}
		// Удваиваем, пока не упёрлись в потолок: иначе через ~34 шага
		// Duration переполнится и станет отрицательной.
		if delay < p.Max {
			delay *= 2
		}

		if err := sleep(ctx, wait); err != nil {
			return fmt.Errorf("повтор прерван: %w (последняя ошибка: %w)", err, last)
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("повтор прерван: %w (последняя ошибка: %w)", err, last)
		}
	}
}

// isTemporary смотрит на поведение, а не на конкретный тип: так же работает
// net.Error. context.DeadlineExceeded тоже отвечает Temporary() == true.
func isTemporary(err error) bool {
	var t interface{ Temporary() bool }
	return errors.As(err, &t) && t.Temporary()
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
