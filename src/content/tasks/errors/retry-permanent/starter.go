package main

import "errors"

// ErrPermanent помечает ошибки, которые повторять бессмысленно.
var ErrPermanent = errors.New("постоянная ошибка")

// Retry вызывает f до attempts раз, пока она не вернёт nil.
//   - Если ошибка оборачивает ErrPermanent — сразу вернуть её, без повторов.
//   - Если попытки кончились — вернуть ошибку, которая оборачивает последнюю
//     ошибку f и упоминает число попыток.
func Retry(attempts int, f func() error) error {
	// ваш код
	return f()
}
