package main

// Filter возвращает новый слайс из тех элементов, для которых keep вернула true.
//
// Порядок сохраняется. Исходный слайс менять нельзя.
// Работать должно с любым типом элементов, а не только с числами.
func Filter[T any](items []T, keep func(T) bool) []T {
	// any, а не comparable и не constraints.Ordered: сравнивать элементы
	// мы не собираемся, решение принимает переданная функция.
	out := make([]T, 0, len(items))
	for _, item := range items {
		if keep(item) {
			out = append(out, item)
		}
	}
	return out
}
