package main

import "slices"

// Rotate сдвигает элементы на k позиций влево на месте:
// [1 2 3 4 5], k=2 → [3 4 5 1 2].
// k может быть больше длины и отрицательным — тогда сдвиг вправо.
func Rotate(xs []int, k int) {
	n := len(xs)
	if n == 0 {
		return
	}
	// Приводим k к диапазону [0, n): остаток в Go бывает отрицательным.
	k = ((k % n) + n) % n
	// Три разворота: обе части по отдельности, потом всё целиком.
	slices.Reverse(xs[:k])
	slices.Reverse(xs[k:])
	slices.Reverse(xs)
}
