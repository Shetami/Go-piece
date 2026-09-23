package main

import (
	"errors"
	"fmt"
)

// ErrPermanent помечает ошибки, которые повторять бессмысленно.
var ErrPermanent = errors.New("постоянная ошибка")

// Retry вызывает f до attempts раз, пока она не вернёт nil.
//   - Если ошибка оборачивает ErrPermanent — сразу вернуть её, без повторов.
//   - Если попытки кончились — вернуть ошибку, которая оборачивает последнюю
//     ошибку f и упоминает число попыток.
func Retry(attempts int, f func() error) error {
	var err error
	for range attempts {
		err = f()
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrPermanent) {
			return err
		}
	}
	return fmt.Errorf("после %d попыток: %w", attempts, err)
}
