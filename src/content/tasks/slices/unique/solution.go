package main

// Unique возвращает элементы без повторов в порядке первого появления.
// Исходный слайс не меняется.
func Unique[T comparable](xs []T) []T {
	// Мапа — множество уже встреченных; struct{} ничего не весит.
	seen := make(map[T]struct{}, len(xs))
	out := make([]T, 0, len(xs))
	for _, x := range xs {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}
