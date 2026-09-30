package main

import (
	"errors"
	"fmt"
)

// ErrPanic помечает ошибки, в которые превратилась паника.
var ErrPanic = errors.New("паника")

// Result — значение или ошибка: так результаты едут по каналам конвейера.
type Result[T any] struct {
	Val T
	Err error
}

// Try вызывает f и упаковывает результат. Паника внутри f превращается в
// Result с ошибкой, для которой errors.Is(err, ErrPanic), а текст содержит
// значение паники; если паниковали ошибкой — errors.Is находит и её.
func Try[T any](f func() (T, error)) (r Result[T]) {
	defer func() {
		if p := recover(); p != nil {
			// Именованный результат: только так defer может его подменить.
			if e, ok := p.(error); ok {
				r = Result[T]{Err: fmt.Errorf("%w: %w", ErrPanic, e)}
			} else {
				r = Result[T]{Err: fmt.Errorf("%w: %v", ErrPanic, p)}
			}
		}
	}()
	v, err := f()
	return Result[T]{v, err}
}

// Map применяет f к значению успешного r. Если в r ошибка, f не
// вызывается, и ошибка переносится как есть.
func Map[T, U any](r Result[T], f func(T) (U, error)) Result[U] {
	if r.Err != nil {
		return Result[U]{Err: r.Err}
	}
	v, err := f(r.Val)
	return Result[U]{v, err}
}

// Collect возвращает все значения, если ошибок нет (для пустого входа —
// пустой, но не nil слайс: в JSON это [], а не null). Иначе — nil и
// errors.Join всех ошибок в исходном порядке, каждая в виде
// "#<индекс>: <текст>" с сохранением цепочки.
func Collect[T any](rs []Result[T]) ([]T, error) {
	var errs []error
	for i, r := range rs {
		if r.Err != nil {
			errs = append(errs, fmt.Errorf("#%d: %w", i, r.Err))
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	vals := make([]T, 0, len(rs))
	for _, r := range rs {
		vals = append(vals, r.Val)
	}
	return vals, nil
}
