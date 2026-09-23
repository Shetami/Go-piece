package main

// GroupBy раскладывает элементы по ключу key.
// Порядок внутри группы — как во входе. Результат не nil даже для пустого входа.
func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, x := range xs {
		k := key(x)
		// append к отсутствующему ключу работает: groups[k] для него — nil.
		out[k] = append(out[k], x)
	}
	return out
}
