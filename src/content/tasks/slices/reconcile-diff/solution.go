package main

import "slices"

// Diff сравнивает желаемое состояние want с текущим have и говорит, что
// добавить (есть в want, нет в have) и что удалить (есть в have, нет в want).
// Равенство — по cmp (cmp(a, b) == 0), а не по ==. Входы — в любом порядке
// и с повторами; входы не меняются. add и remove отсортированы по cmp и без
// повторов; из нескольких равных по cmp элементов остаётся первый по порядку
// во входе. Время — O((n+m) log(n+m)).
func Diff[T any](want, have []T, cmp func(a, b T) int) (add, remove []T) {
	w, h := sortedUnique(want, cmp), sortedUnique(have, cmp)
	i, j := 0, 0
	for i < len(w) && j < len(h) {
		switch c := cmp(w[i], h[j]); {
		case c < 0: // w[i] меньше всех оставшихся h — в have его нет
			add = append(add, w[i])
			i++
		case c > 0:
			remove = append(remove, h[j])
			j++
		default: // есть в обоих — ничего делать не надо
			i++
			j++
		}
	}
	add = append(add, w[i:]...)
	remove = append(remove, h[j:]...)
	return add, remove
}

// sortedUnique — отсортированная копия без повторов по cmp.
func sortedUnique[T any](s []T, cmp func(a, b T) int) []T {
	c := slices.Clone(s)
	// Стабильная сортировка: среди равных первым остаётся первый во входе,
	// и CompactFunc сохраняет именно его.
	slices.SortStableFunc(c, cmp)
	return slices.CompactFunc(c, func(a, b T) bool { return cmp(a, b) == 0 })
}
