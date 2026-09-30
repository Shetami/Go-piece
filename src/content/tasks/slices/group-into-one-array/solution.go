package main

// Group — ключ и все элементы с этим ключом.
type Group[K comparable, T any] struct {
	Key   K
	Items []T
}

// GroupBy раскладывает s по группам ключа key. Группы идут в порядке первого
// появления ключа, элементы внутри группы — в исходном порядке.
// key вызывается ровно один раз на элемент. s не меняется.
//
// Память: элементы всех групп лежат в одном общем массиве длины len(s) —
// одна аллокация под данные, а не рост отдельного слайса у каждой группы.
// При этом append к Items одной группы не должен затирать соседнюю.
func GroupBy[T any, K comparable](s []T, key func(T) K) []Group[K, T] {
	// Проход 1: номер группы каждого элемента и размеры групп.
	gi := make([]int, len(s))
	pos := make(map[K]int)
	var keys []K
	var counts []int
	for i, v := range s {
		k := key(v) // единственный вызов key для элемента
		g, ok := pos[k]
		if !ok {
			g = len(keys)
			pos[k] = g
			keys = append(keys, k)
			counts = append(counts, 0)
		}
		gi[i] = g
		counts[g]++
	}

	// Нарезаем общий массив: у каждой группы len 0 и cap ровно под её элементы.
	backing := make([]T, len(s))
	groups := make([]Group[K, T], len(keys))
	off := 0
	for g, k := range keys {
		groups[g] = Group[K, T]{Key: k, Items: backing[off:off:off+counts[g]]}
		off += counts[g]
	}

	// Проход 2: раскладываем. append не перевыделяет — вместимость точная.
	for i, v := range s {
		groups[gi[i]].Items = append(groups[gi[i]].Items, v)
	}
	return groups
}
