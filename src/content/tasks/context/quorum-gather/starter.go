package main

import (
	"context"
	"errors"
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
	// ваш код
	var got []T
	for _, r := range replicas {
		if v, err := r(ctx); err == nil {
			got = append(got, v)
		}
	}
	return got, nil
}
