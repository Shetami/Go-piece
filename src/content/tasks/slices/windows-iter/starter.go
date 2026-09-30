package main

import "iter"

// Windows возвращает итератор по окнам длины size с шагом step:
// s[0:size], s[step:step+size], s[2*step:2*step+size], …
// Выдаются только полные окна; если len(s) < size — ни одного.
// Окна — подслайсы s без копирования, и append к окну не меняет s.
// Итератор можно обходить повторно, и он корректно останавливается по break.
// size <= 0 или step <= 0 — паника сразу при вызове Windows, а не при обходе.
func Windows[T any](s []T, size, step int) iter.Seq[[]T] {
	// ваш код
	return func(yield func([]T) bool) {}
}
