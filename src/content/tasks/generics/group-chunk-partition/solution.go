package main

// GroupBy раскладывает items по ключам. keys — ключи в порядке первого
// появления, groups[k] — элементы с ключом k в исходном порядке.
// Для пустого items: пустой keys и пустая, но не nil мапа.
func GroupBy[T any, K comparable](items []T, key func(T) K) (keys []K, groups map[K][]T) {
	groups = make(map[K][]T)
	for _, v := range items {
		k := key(v)
		if _, seen := groups[k]; !seen {
			keys = append(keys, k) // порядок мапы случаен — запоминаем свой
		}
		groups[k] = append(groups[k], v)
	}
	return keys, groups
}

// Chunk режет items на куски по size элементов, последний может быть
// короче. Для size <= 0 и пустого items — nil. append к куску не должен
// менять ни соседние куски, ни items.
func Chunk[T any](items []T, size int) [][]T {
	if size <= 0 || len(items) == 0 {
		return nil
	}
	out := make([][]T, 0, (len(items)+size-1)/size)
	for len(items) > 0 {
		n := min(size, len(items))
		// Полное выражение среза: cap == len, append скопирует, а не
		// затрёт начало следующего куска.
		out = append(out, items[:n:n])
		items = items[n:]
	}
	return out
}

// Partition делит items на подходящие под pred и остальные, сохраняя
// порядок. items не меняется.
func Partition[T any](items []T, pred func(T) bool) (yes, no []T) {
	for _, v := range items {
		if pred(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return yes, no
}
