package main

import "errors"

var ErrDuplicate = errors.New("повторяющееся значение")

// Invert меняет местами ключи и значения.
// Если два ключа дают одно значение, возвращает nil и ошибку,
// которая оборачивает ErrDuplicate и называет это значение.
func Invert[K, V comparable](m map[K]V) (map[V]K, error) {
	// ваш код
	return nil, nil
}
