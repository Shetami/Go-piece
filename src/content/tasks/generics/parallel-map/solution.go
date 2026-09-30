package main

import (
	"context"
	"sync"
)

// ParallelMap возвращает out, где out[i] = f(ctx, in[i]), запуская не
// больше limit вызовов f одновременно (limit <= 0 значит 1).
//
// Если какой-то f вернул ошибку, контекст, переданный в f, отменяется,
// новые вызовы не начинаются, ParallelMap дожидается уже запущенных и
// возвращает (nil, ошибку этого f). Если отменён внешний ctx и обход не
// закончен — (nil, ctx.Err()). После возврата не остаётся горутин.
func ParallelMap[T, U any](ctx context.Context, in []T, limit int, f func(context.Context, T) (U, error)) ([]U, error) {
	if limit <= 0 {
		limit = 1
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make([]U, len(in))
	sem := make(chan struct{}, limit) // семафор: занятые слоты = работающие f
	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	stopped := false
	for i, v := range in {
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
		}
		// select выбирает случайно, если готовы обе ветки, — перепроверяем.
		if ctx.Err() != nil {
			stopped = true
			break
		}
		wg.Add(1)
		go func() {
			defer func() { <-sem; wg.Done() }()
			u, err := f(ctx, v)
			if err != nil {
				// Запоминаем только первую ошибку и сразу гасим остальных.
				once.Do(func() { firstErr = err; cancel() })
				return
			}
			out[i] = u // каждая горутина пишет в свой индекс — гонки нет
		}()
	}
	wg.Wait() // без этого горутины переживут вызов и будут писать в out

	if firstErr != nil {
		return nil, firstErr
	}
	if stopped {
		return nil, ctx.Err()
	}
	return out, nil
}
