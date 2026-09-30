package main

import (
	"errors"
	"fmt"
)

// Result — значение или ошибка одного элемента.
type Result[T any] struct {
	Val T
	Err error
}

// ItemError — ошибка элемента с его индексом во входе.
type ItemError struct {
	Index int
	Err   error
}

// Error возвращает "элемент <Index>: <текст Err>".
func (e *ItemError) Error() string { return fmt.Sprintf("элемент %d: %v", e.Index, e.Err) }

// Unwrap открывает Err для errors.Is и errors.As.
func (e *ItemError) Unwrap() error { return e.Err }

// Collect разбирает результаты: значения успешных — в исходном порядке,
// ошибки — каждая обёрнута в *ItemError с индексом в rs и все
// объединены errors.Join. Если ошибок нет, err == nil.
func Collect[T any](rs []Result[T]) ([]T, error) {
	var (
		vals []T
		errs []error // именно []error: nil-интерфейсы Join выбросит сам
	)
	for i, r := range rs {
		if r.Err != nil {
			errs = append(errs, &ItemError{Index: i, Err: r.Err})
			continue
		}
		vals = append(vals, r.Val)
	}
	// Join без ошибок возвращает настоящий nil, а не пустую обёртку.
	return vals, errors.Join(errs...)
}

// TryMap применяет f к каждому элементу in — даже после ошибок — и
// возвращает то же, что Collect от результатов.
func TryMap[T, U any](in []T, f func(T) (U, error)) ([]U, error) {
	rs := make([]Result[U], len(in))
	for i, v := range in {
		rs[i].Val, rs[i].Err = f(v)
	}
	return Collect(rs)
}
