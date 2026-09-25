package main

import "cmp"

// MaxBy возвращает элемент s с наибольшим ключом key(x) и true.
// При равных ключах — первый из них. Для пустого s — нулевое значение и false.
// key вызывается не больше одного раза на элемент.
func MaxBy[T any, K cmp.Ordered](s []T, key func(T) K) (T, bool) {
	// ваш код
	var zero T
	return zero, false
}
