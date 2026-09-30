package main

import "log/slog"

// WithFields прикрепляет к ошибке поля для структурного лога, не меняя её
// текст и не мешая errors.Is/As. kv — чередование ключ (string), значение.
// nil → nil.
func WithFields(err error, kv ...any) error {
	// ваш код
	return err
}

// Fields собирает поля со всего дерева err (цепочки Unwrap() error и все
// ветки Unwrap() []error). При совпадении ключей побеждает поле,
// прикреплённое ГЛУБЖЕ (ближе к источнику ошибки), а между ветками одного
// Join — более ранняя ветка. Нет полей — пустая, но не nil мапа.
func Fields(err error) map[string]any {
	// ваш код
	return nil
}

// Attrs — поля для slog: отсортированы по ключу, последним добавлен
// атрибут "err" с текстом ошибки. nil → nil.
func Attrs(err error) []slog.Attr {
	// ваш код
	return nil
}
