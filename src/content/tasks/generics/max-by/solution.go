package main

import "cmp"

// MaxBy возвращает элемент s с наибольшим ключом key(x) и true.
// При равных ключах — первый из них. Для пустого s — нулевое значение и false.
// key вызывается не больше одного раза на элемент.
func MaxBy[T any, K cmp.Ordered](s []T, key func(T) K) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	best, bestKey := s[0], key(s[0])
	for _, x := range s[1:] {
		// Ключ лучшего запомнен — не пересчитываем его на каждом шаге.
		// Строгое > оставляет первый из равных.
		if k := key(x); k > bestKey {
			best, bestKey = x, k
		}
	}
	return best, true
}
