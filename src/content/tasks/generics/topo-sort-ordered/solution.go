package main

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

// ErrCycle — в зависимостях есть цикл.
var ErrCycle = errors.New("цикл в зависимостях")

// TopoSort упорядочивает узлы так, что каждый идёт после всех, от кого
// зависит: deps[n] — список зависимостей n. Узлы, которые встречаются
// только в списках зависимостей, тоже входят в ответ. Из нескольких
// готовых узлов первым берётся наименьший — ответ детерминирован.
// При цикле — (nil, ошибка), для которой errors.Is(err, ErrCycle), с
// текстом "цикл в зависимостях: [узлы]", где перечислены по возрастанию
// все узлы, которые не удалось упорядочить.
func TopoSort[N cmp.Ordered](deps map[N][]N) ([]N, error) {
	indeg := map[N]int{} // сколько зависимостей ещё не выписано
	users := map[N][]N{} // кто зависит от узла
	for n, ds := range deps {
		if _, ok := indeg[n]; !ok {
			indeg[n] = 0
		}
		for _, d := range ds {
			if _, ok := indeg[d]; !ok {
				indeg[d] = 0 // узел только из списка зависимостей
			}
			indeg[n]++ // повторная зависимость считается и снимается дважды
			users[d] = append(users[d], n)
		}
	}

	var ready []N // отсортирован: берём наименьший с начала
	for n, k := range indeg {
		if k == 0 {
			ready = append(ready, n)
		}
	}
	slices.Sort(ready)

	out := make([]N, 0, len(indeg))
	for len(ready) > 0 {
		n := ready[0]
		ready = ready[1:]
		out = append(out, n)
		for _, u := range users[n] {
			indeg[u]--
			if indeg[u] == 0 {
				i, _ := slices.BinarySearch(ready, u)
				ready = slices.Insert(ready, i, u)
			}
		}
	}

	if len(out) < len(indeg) {
		var rest []N
		for n, k := range indeg {
			if k > 0 {
				rest = append(rest, n)
			}
		}
		slices.Sort(rest)
		return nil, fmt.Errorf("%w: %v", ErrCycle, rest)
	}
	return out, nil
}
