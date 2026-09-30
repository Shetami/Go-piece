package main

import (
	"context"
	"sync"
)

// RunPriority запускает workers воркеров, которые выполняют задачи из двух
// очередей: high и low.
//
//   - Если в high есть задача, воркер берёт её раньше любой задачи из low.
//   - Закрытие одной очереди не останавливает работу с другой; RunPriority
//     возвращает nil, когда обе очереди закрыты и все задачи выполнены.
//   - При отмене ctx воркеры не начинают новых задач; RunPriority
//     дожидается уже выполняющихся и возвращает ctx.Err().
func RunPriority(ctx context.Context, workers int, high, low <-chan func()) error {
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, high, low)
		}()
	}
	wg.Wait()
	return ctx.Err()
}

func worker(ctx context.Context, high, low <-chan func()) {
	// Локальные копии: закрытую очередь обнуляем, и select её больше не видит.
	for high != nil || low != nil {
		if ctx.Err() != nil {
			return
		}
		// 1. Сначала — только high, без ожидания.
		select {
		case task, ok := <-high:
			if !ok {
				high = nil
				continue
			}
			task()
			continue
		default:
		}
		// 2. high пуст — ждём что угодно. Если придёт high, возьмём его.
		select {
		case task, ok := <-high:
			if !ok {
				high = nil
				continue
			}
			task()
		case task, ok := <-low:
			if !ok {
				low = nil
				continue
			}
			task()
		case <-ctx.Done():
			return
		}
	}
}
