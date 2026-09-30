package main

import (
	"context"
	"errors"
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
	// ваш код
	return f(ctx)
}
