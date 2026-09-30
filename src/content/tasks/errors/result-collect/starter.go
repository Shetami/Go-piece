package main

import "errors"

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
func Try[T any](f func() (T, error)) Result[T] {
	// ваш код
	v, err := f()
	return Result[T]{v, err}
}

// Map применяет f к значению успешного r. Если в r ошибка, f не
// вызывается, и ошибка переносится как есть.
func Map[T, U any](r Result[T], f func(T) (U, error)) Result[U] {
	// ваш код
	return Result[U]{}
}

// Collect возвращает все значения, если ошибок нет (для пустого входа —
// пустой, но не nil слайс: в JSON это [], а не null). Иначе — nil и
// errors.Join всех ошибок в исходном порядке, каждая в виде
// "#<индекс>: <текст>" с сохранением цепочки.
func Collect[T any](rs []Result[T]) ([]T, error) {
	// ваш код
	return nil, nil
}
