package main

// GroupBy раскладывает элементы по ключу key.
// Порядок внутри группы — как во входе. Результат не nil даже для пустого входа.
func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	// ваш код
	return nil
}
