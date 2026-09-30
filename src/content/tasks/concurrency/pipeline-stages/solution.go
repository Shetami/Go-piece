package main

import (
	"context"
	"sync"
)

// Stage — один этап конвейера.
type Stage[T any] func(ctx context.Context, v T) (T, error)

// Run прогоняет каждое значение из src через все stages по порядку.
//
// Каждый этап работает в своей горутине, соседние этапы соединены каналами:
// пока этап 2 обрабатывает значение i, этап 1 уже может взять i+1.
// Результаты — в порядке src. Без этапов — копия src.
//
// Первая ошибка любого этапа останавливает конвейер: Run возвращает nil и
// именно эту ошибку (а не context.Canceled, который после неё могут вернуть
// другие этапы). Отмена ctx снаружи — nil и ctx.Err().
// К моменту возврата Run все её горутины завершены.
func Run[T any](ctx context.Context, src []T, stages ...Stage[T]) ([]T, error) {
	// Cause запомнит первую причину отмены — ошибку этапа.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var wg sync.WaitGroup
	source := make(chan T)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(source) // закрывает тот, кто пишет
		for _, v := range src {
			select {
			case source <- v:
			case <-ctx.Done():
				return
			}
		}
	}()

	var in <-chan T = source
	for _, st := range stages {
		out := make(chan T)
		wg.Add(1)
		go func(in <-chan T) {
			defer wg.Done()
			defer close(out)
			for v := range in {
				r, err := st(ctx, v)
				if err != nil {
					cancel(err) // будит всех, кто застрял на отправке
					return
				}
				select {
				case out <- r:
				case <-ctx.Done():
					return
				}
			}
		}(in)
		in = out
	}

	res := make([]T, 0, len(src))
	for v := range in { // закроется, когда закроется последний этап
		res = append(res, v)
	}
	wg.Wait()
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	return res, nil
}
