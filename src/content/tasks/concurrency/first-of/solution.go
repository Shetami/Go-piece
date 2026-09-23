package main

import (
	"context"
	"errors"
)

// FirstOf запускает все функции параллельно и возвращает первый успешный
// результат. Как только он получен, контекст остальных функций отменяется.
// Если упали все — ошибка, в которой есть ошибки каждой.
func FirstOf(ctx context.Context, fns ...func(context.Context) (string, error)) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // выход с результатом отменяет всех оставшихся

	type result struct {
		val string
		err error
	}
	// Буфер на всех: опоздавшие отправят результат и завершатся,
	// даже когда их уже никто не читает.
	results := make(chan result, len(fns))
	for _, fn := range fns {
		go func() {
			v, err := fn(ctx)
			results <- result{v, err}
		}()
	}

	var errs []error
	for range fns {
		r := <-results
		if r.err == nil {
			return r.val, nil
		}
		errs = append(errs, r.err)
	}
	return "", errors.Join(errs...)
}
