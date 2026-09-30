package main

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

// ParallelMap применяет f к каждому элементу in и возвращает результаты в
// исходном порядке. Одновременно работает не больше runtime.GOMAXPROCS(0)
// вызовов f — столько, сколько у процесса «процессоров» Go (для
// CPU-bound работы больше не имеет смысла). Сама функция GOMAXPROCS не меняет.
//
// При первой ошибке f: новые элементы не запускаются, ctx, переданный в
// уже работающие f, отменяется, ParallelMap дожидается их и возвращает
// именно эту первую ошибку (а не context.Canceled от отменённых). Если
// отменили родительский ctx — возвращается ctx.Err(). После возврата
// не остаётся ни одной запущенной горутины.
func ParallelMap[T, R any](ctx context.Context, in []T, f func(context.Context, T) (R, error)) ([]R, error) {
	out := make([]R, len(in))
	// GOMAXPROCS(0) только читает значение; с аргументом > 0 — меняет
	// его для всего процесса.
	workers := min(runtime.GOMAXPROCS(0), len(in))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		next     atomic.Int64 // индекс следующего элемента
		errOnce  sync.Once
		firstErr error
		wg       sync.WaitGroup
	)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				i := int(next.Add(1) - 1)
				if i >= len(in) {
					return
				}
				r, err := f(ctx, in[i])
				if err != nil {
					// Запоминаем только первую ошибку: остальные — чаще
					// всего context.Canceled как следствие нашей отмены.
					errOnce.Do(func() { firstErr = err; cancel() })
					return
				}
				out[i] = r // каждый индекс пишет ровно одна горутина — гонки нет
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		// Отмена родителя (своя отмена случается только вместе с firstErr).
		return nil, err
	}
	return out, nil
}
