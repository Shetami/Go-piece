package main

import "iter"

// Windows возвращает итератор по окнам длины size с шагом step:
// s[0:size], s[step:step+size], s[2*step:2*step+size], …
// Выдаются только полные окна; если len(s) < size — ни одного.
// Окна — подслайсы s без копирования, и append к окну не меняет s.
// Итератор можно обходить повторно, и он корректно останавливается по break.
// size <= 0 или step <= 0 — паника сразу при вызове Windows, а не при обходе.
func Windows[T any](s []T, size, step int) iter.Seq[[]T] {
	// Проверяем здесь: тело итератора выполнится, только когда его начнут обходить.
	if size <= 0 || step <= 0 {
		panic("Windows: size и step должны быть больше нуля")
	}
	return func(yield func([]T) bool) {
		for i := 0; i+size <= len(s); i += step {
			// Третий индекс: append к окну не пишет в следующий элемент s.
			if !yield(s[i : i+size : i+size]) {
				return // break в цикле range — больше yield звать нельзя
			}
		}
	}
}
