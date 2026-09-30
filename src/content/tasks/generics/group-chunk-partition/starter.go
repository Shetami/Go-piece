package main

// GroupBy раскладывает items по ключам. keys — ключи в порядке первого
// появления, groups[k] — элементы с ключом k в исходном порядке.
// Для пустого items: пустой keys и пустая, но не nil мапа.
func GroupBy[T any, K comparable](items []T, key func(T) K) (keys []K, groups map[K][]T) {
	// ваш код
	return nil, nil
}

// Chunk режет items на куски по size элементов, последний может быть
// короче. Для size <= 0 и пустого items — nil. append к куску не должен
// менять ни соседние куски, ни items.
func Chunk[T any](items []T, size int) [][]T {
	// ваш код
	return nil
}

// Partition делит items на подходящие под pred и остальные, сохраняя
// порядок. items не меняется.
func Partition[T any](items []T, pred func(T) bool) (yes, no []T) {
	// ваш код
	return nil, nil
}
