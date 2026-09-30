package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrNoReplicas = errors.New("hedged: нет реплик")

// Replica — запрос к одной реплике. Обязана слушать ctx.
type Replica func(ctx context.Context) (string, error)

// Hedged — запрос «с подстраховкой». Сначала запускается replicas[0].
// Если за delay ответа нет — запускается следующая реплика, ещё через
// delay — следующая, и так далее; предыдущие при этом продолжают работать.
// Если реплика вернула ошибку — следующая запускается сразу, не дожидаясь
// delay. Первый успешный ответ возвращается немедленно, а контекст всех
// остальных запущенных реплик отменяется.
// Все реплики упали — ошибка errors.Join всех их ошибок (каждая с номером
// реплики). Отменили ctx — context.Cause(ctx) сразу.
// Горутины реплик не должны оставаться висеть после выхода.
func Hedged(ctx context.Context, delay time.Duration, replicas []Replica) (string, error) {
	if len(replicas) == 0 {
		return "", ErrNoReplicas
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // отмена проигравших — при любом выходе

	type result struct {
		i   int
		v   string
		err error
	}
	// Буфер на всех: проигравшим есть куда записать ответ, даже когда его уже никто не читает.
	results := make(chan result, len(replicas))
	next, running := 0, 0
	launch := func() {
		i := next
		next++
		running++
		go func() {
			v, err := replicas[i](ctx)
			results <- result{i, v, err}
		}()
	}

	launch()
	timer := time.NewTimer(delay)
	defer timer.Stop()
	var errs []error
	for {
		select {
		case r := <-results:
			running--
			if r.err == nil {
				return r.v, nil
			}
			errs = append(errs, fmt.Errorf("реплика %d: %w", r.i, r.err))
			if next < len(replicas) {
				launch() // упала — подстраховку не ждём
				timer.Reset(delay)
			} else if running == 0 {
				return "", errors.Join(errs...)
			}
		case <-timer.C:
			if next < len(replicas) {
				launch()
				timer.Reset(delay)
			}
		case <-ctx.Done():
			return "", context.Cause(ctx)
		}
	}
}
