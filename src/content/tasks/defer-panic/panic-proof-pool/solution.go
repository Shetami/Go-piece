package main

import (
	"context"
	"fmt"
	"sync"
)

// Result — результат обработки одного элемента.
type Result[R any] struct {
	Value R
	Err   error
}

// Map обрабатывает items не более чем в workers горутинах одновременно
// (workers < 1 считается за 1) и возвращает results, где results[i]
// соответствует items[i].
//
//   - Паника fn на элементе становится ошибкой этого элемента: в тексте есть
//     значение паники, а если паниковали ошибкой — errors.Is находит её.
//     Воркер после паники продолжает обрабатывать следующие элементы.
//   - Если ctx отменён, fn для ещё не начатых элементов не вызывается,
//     их Err = ctx.Err().
//   - Map возвращается, только когда все её горутины завершились.
func Map[T, R any](ctx context.Context, items []T, workers int, fn func(context.Context, T) (R, error)) []Result[R] {
	workers = max(workers, 1)
	results := make([]Result[R], len(items))
	jobs := make(chan int)

	var wg sync.WaitGroup
	for range min(workers, len(items)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if err := ctx.Err(); err != nil {
					results[i].Err = err
					continue
				}
				// Каждый элемент — в своём вызове с recover: паника одной
				// задачи не убивает воркера и не оставляет пул без рабочих.
				results[i] = runItem(ctx, items[i], fn)
			}
		}()
	}

	for i := range items {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

// runItem вызывает fn и превращает её панику в ошибку результата.
func runItem[T, R any](ctx context.Context, item T, fn func(context.Context, T) (R, error)) (res Result[R]) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				res = Result[R]{Err: fmt.Errorf("паника в задаче: %w", err)}
			} else {
				res = Result[R]{Err: fmt.Errorf("паника в задаче: %v", r)}
			}
		}
	}()
	v, err := fn(ctx, item)
	return Result[R]{Value: v, Err: err}
}
