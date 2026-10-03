package main

import (
	"errors"
	"slices"
)

// FewestCoins возвращает самый короткий набор монет из coins (каждого
// номинала сколько угодно), который в сумме даёт target, — по возрастанию.
// Ошибка — если target отрицательный или собрать его нельзя.
func FewestCoins(coins []int, target int) ([]int, error) {
	if target < 0 {
		return nil, errors.New("target can't be negative")
	}
	// fewest[s] — сколько монет нужно на сумму s, -1 — собрать нельзя;
	// last[s] — какой монетой эта сумма закрыта в лучшем наборе.
	fewest := make([]int, target+1)
	last := make([]int, target+1)
	for s := 1; s <= target; s++ {
		fewest[s] = -1
		for _, c := range coins {
			if c > s || fewest[s-c] < 0 {
				continue
			}
			if n := fewest[s-c] + 1; fewest[s] < 0 || n < fewest[s] {
				fewest[s], last[s] = n, c
			}
		}
	}
	if fewest[target] < 0 {
		return nil, errors.New("can't make target with given coins")
	}

	// Восстанавливаем набор, откатываясь от target по last.
	out := make([]int, 0, fewest[target])
	for s := target; s > 0; s -= last[s] {
		out = append(out, last[s])
	}
	slices.Sort(out)
	return out, nil
}
