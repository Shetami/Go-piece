package main

import (
	"errors"
	"slices"
)

// Rebase переводит число, записанное цифрами digits в системе счисления
// inputBase, в систему outputBase. Старшая цифра — первая.
// Ошибка — если какое-то основание меньше 2 или цифра вне [0, inputBase).
func Rebase(inputBase int, digits []int, outputBase int) ([]int, error) {
	if inputBase < 2 {
		return nil, errors.New("input base must be >= 2")
	}
	if outputBase < 2 {
		return nil, errors.New("output base must be >= 2")
	}

	// Схема Горнера: n = (((d0·b + d1)·b + d2)·b + ...).
	n := 0
	for _, d := range digits {
		if d < 0 || d >= inputBase {
			return nil, errors.New("all digits must satisfy 0 <= d < input base")
		}
		n = n*inputBase + d
	}
	if n == 0 {
		return []int{0}, nil
	}

	// Остатки от деления дают цифры с младшей — разворачиваем в конце.
	var out []int
	for ; n > 0; n /= outputBase {
		out = append(out, n%outputBase)
	}
	slices.Reverse(out)
	return out, nil
}
