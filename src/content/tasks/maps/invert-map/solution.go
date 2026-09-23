package main

import (
	"errors"
	"fmt"
)

var ErrDuplicate = errors.New("повторяющееся значение")

// Invert меняет местами ключи и значения.
// Если два ключа дают одно значение, возвращает nil и ошибку,
// которая оборачивает ErrDuplicate и называет это значение.
func Invert[K, V comparable](m map[K]V) (map[V]K, error) {
	out := make(map[V]K, len(m))
	for k, v := range m {
		if _, ok := out[v]; ok {
			// %w — чтобы errors.Is(err, ErrDuplicate) работал у вызывающего.
			return nil, fmt.Errorf("%w: %v", ErrDuplicate, v)
		}
		out[v] = k
	}
	return out, nil
}
