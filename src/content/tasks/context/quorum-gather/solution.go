package main

import (
	"context"
	"errors"
	"fmt"
)

var ErrNoQuorum = errors.New("кворум недостижим")

// Quorum параллельно вызывает все replicas и возвращает первые k успешных
// ответов (в порядке прихода), как только их набралось k. Контекст
// остальных вызовов при этом отменяется.
//   - если кворум уже не набрать (ошибок больше, чем len(replicas)-k),
//     возвращается сразу, не дожидаясь оставшихся: ошибка отвечает
//     errors.Is на ErrNoQuorum и на каждую полученную ошибку реплик;
//   - k < 1 или k > len(replicas) — ErrNoQuorum без единого вызова;
//   - отмена ctx — context.Cause(ctx) сразу.
//
// Горутины вызовов не остаются висеть после выхода (если реплики слушают ctx).
func Quorum[T any](ctx context.Context, k int, replicas []func(context.Context) (T, error)) ([]T, error) {
	n := len(replicas)
	if k < 1 || k > n {
		return nil, fmt.Errorf("нужно %d из %d: %w", k, n, ErrNoQuorum)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // отменяем тех, кто ещё работает, при любом выходе

	type result struct {
		v   T
		err error
	}
	results := make(chan result, n) // буфер: опоздавшим есть куда писать
	for _, r := range replicas {
		go func() {
			v, err := r(ctx)
			results <- result{v, err}
		}()
	}

	var got []T
	var errs []error
	for range n {
		select {
		case r := <-results:
			if r.err != nil {
				errs = append(errs, r.err)
				if len(errs) > n-k { // даже если все оставшиеся ответят — не хватит
					return nil, errors.Join(append([]error{ErrNoQuorum}, errs...)...)
				}
				continue
			}
			if got = append(got, r.v); len(got) == k {
				return got, nil
			}
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		}
	}
	return nil, errors.Join(append([]error{ErrNoQuorum}, errs...)...) // сюда не дойдём
}
