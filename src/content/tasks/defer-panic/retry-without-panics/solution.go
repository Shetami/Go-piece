package main

import (
	"context"
	"errors"
	"fmt"
)

// PanicError — паника fn, превращённая в ошибку.
type PanicError struct{ Value any }

func (e *PanicError) Error() string { return fmt.Sprintf("паника: %v", e.Value) }

// Unwrap отдаёт исходную ошибку, если паниковали ошибкой.
func (e *PanicError) Unwrap() error { err, _ := e.Value.(error); return err }

// Retry вызывает fn(ctx), пока она возвращает ошибку, но не больше
// attempts раз (attempts < 1 — один раз).
//   - Успех — nil.
//   - Между попытками (но не после последней) вызывается wait(ctx, n),
//     где n — номер только что неудавшейся попытки, начиная с 1.
//     Если wait вернул ошибку (контекст отменён) — Retry прекращается и
//     возвращает errors.Join(ошибка последней попытки, ошибка wait).
//   - Паника fn не повторяется: Retry сразу возвращает *PanicError.
//   - Попытки кончились — вернуть ошибку последней попытки.
func Retry(ctx context.Context, attempts int, wait func(ctx context.Context, n int) error, fn func(ctx context.Context) error) error {
	attempts = max(attempts, 1)
	var err error
	for n := 1; n <= attempts; n++ {
		var panicked bool
		err, panicked = callOnce(ctx, fn)
		if err == nil || panicked {
			return err // паника — баг, а не временный сбой: повтор не поможет
		}
		if n == attempts {
			break
		}
		if werr := wait(ctx, n); werr != nil {
			return errors.Join(err, werr)
		}
	}
	return err
}

// callOnce отделяет «fn вернула ошибку» от «fn запаниковала».
func callOnce(ctx context.Context, fn func(context.Context) error) (err error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			err, panicked = &PanicError{Value: r}, true
		}
	}()
	return fn(ctx), false
}
