package main

import "context"

// FirstOf запускает все функции параллельно и возвращает первый успешный
// результат. Как только он получен, контекст остальных функций отменяется.
// Если упали все — ошибка, в которой есть ошибки каждой.
func FirstOf(ctx context.Context, fns ...func(context.Context) (string, error)) (string, error) {
	// ваш код
	return "", nil
}
