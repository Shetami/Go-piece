package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Policy — настройки повторов.
type Policy struct {
	Attempts int           // максимум попыток, >= 1
	Base     time.Duration // пауза перед второй попыткой
	Max      time.Duration // потолок паузы
	// Jitter выбирает фактическую паузу из [0, d]. nil — пауза ровно d.
	Jitter func(d time.Duration) time.Duration
}

// RetryAfterError — ошибка, которая сама говорит, сколько ждать (как Retry-After).
type RetryAfterError interface {
	error
	RetryAfter() time.Duration
}

var ErrNoTime = errors.New("retry: до дедлайна не хватит времени на паузу")

// Do вызывает f, пока она не вернёт nil, но не больше p.Attempts раз.
// Пауза перед попыткой k+1 (k = 1, 2, …) — min(Base·2^(k-1), Max), затем
// через Jitter. Если последняя ошибка (через errors.As) — RetryAfterError,
// пауза равна её RetryAfter(), без Jitter и без потолка.
// Если у ctx есть дедлайн и пауза закончится позже него — Do не ждёт:
// сразу возвращает ошибку, отвечающую errors.Is на ErrNoTime и на последнюю
// ошибку f. Отмена ctx (до попытки или во время паузы) — ошибка,
// отвечающая errors.Is на context.Cause(ctx) и на последнюю ошибку f (если была).
// Попытки кончились — последняя ошибка f как есть.
func Do(ctx context.Context, p Policy, f func(context.Context) error) error {
	var last error
	for k := 1; ; k++ {
		if ctx.Err() != nil {
			return fmt.Errorf("retry: попытка %d отменена: %w", k, errors.Join(context.Cause(ctx), last))
		}
		if last = f(ctx); last == nil {
			return nil
		}
		if k >= p.Attempts {
			return last
		}

		pause := p.backoff(k)
		var ra RetryAfterError
		if errors.As(last, &ra) {
			pause = ra.RetryAfter() // сервер знает лучше нас
		} else if p.Jitter != nil {
			pause = p.Jitter(pause)
		}

		// Бессмысленно спать, если проснёмся уже после дедлайна.
		if dl, ok := ctx.Deadline(); ok && time.Now().Add(pause).After(dl) {
			return errors.Join(ErrNoTime, last)
		}

		t := time.NewTimer(pause)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return fmt.Errorf("retry: пауза после попытки %d прервана: %w", k, errors.Join(context.Cause(ctx), last))
		}
	}
}

// backoff — Base·2^(k-1), но не больше Max и без переполнения.
func (p Policy) backoff(k int) time.Duration {
	d := p.Base
	for i := 1; i < k; i++ {
		if d >= p.Max/2 { // следующее удвоение перешагнёт потолок (или переполнится)
			return p.Max
		}
		d *= 2
	}
	return min(d, p.Max)
}
