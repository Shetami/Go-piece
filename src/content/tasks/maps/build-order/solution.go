package main

import (
	"container/heap"
	"errors"
	"slices"
	"strings"
)

var ErrCycle = errors.New("цикл зависимостей")

// CycleError — ошибка BuildOrder при цикле. errors.Is(err, ErrCycle) == true.
// Nodes — все вершины, которые не удалось упорядочить (лежат на цикле или
// зависят от него), по алфавиту.
type CycleError struct {
	Nodes []string
}

func (e *CycleError) Error() string {
	return "цикл зависимостей среди: " + strings.Join(e.Nodes, ", ")
}

// Unwrap связывает CycleError с ErrCycle.
func (e *CycleError) Unwrap() error { return ErrCycle }

// strHeap — минимальная куча строк: готовые к сборке, по алфавиту.
type strHeap []string

func (h strHeap) Len() int           { return len(h) }
func (h strHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h strHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *strHeap) Push(x any)        { *h = append(*h, x.(string)) }
func (h *strHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

// BuildOrder возвращает порядок сборки: deps[x] — от чего зависит x, и x
// должен идти после всех своих зависимостей.
//
//   - Вершины, которые встречаются только в списках зависимостей, тоже
//     входят в ответ.
//   - Повтор зависимости в списке ("a": {"b", "b"}) — одна зависимость.
//   - Порядок однозначен: из всех вершин, готовых к сборке, первой берётся
//     наименьшая по алфавиту.
//   - При цикле — nil и *CycleError.
func BuildOrder(deps map[string][]string) ([]string, error) {
	indeg := make(map[string]int)      // сколько зависимостей ещё не собрано
	users := make(map[string][]string) // зависимость → кто от неё зависит
	seen := make(map[[2]string]bool)   // рёбра без повторов
	for x, ds := range deps {
		if _, ok := indeg[x]; !ok {
			indeg[x] = 0
		}
		for _, d := range ds {
			if _, ok := indeg[d]; !ok {
				indeg[d] = 0 // вершина только из списка зависимостей
			}
			if seen[[2]string{x, d}] {
				continue
			}
			seen[[2]string{x, d}] = true
			indeg[x]++
			users[d] = append(users[d], x)
		}
	}

	var ready strHeap
	for x, n := range indeg {
		if n == 0 {
			ready = append(ready, x)
		}
	}
	heap.Init(&ready)

	order := make([]string, 0, len(indeg))
	for ready.Len() > 0 {
		x := heap.Pop(&ready).(string)
		order = append(order, x)
		for _, u := range users[x] {
			indeg[u]--
			if indeg[u] == 0 {
				heap.Push(&ready, u)
			}
		}
	}

	if len(order) < len(indeg) {
		// Кто остался с ненулевой степенью — на цикле или за ним.
		var rest []string
		for x, n := range indeg {
			if n > 0 {
				rest = append(rest, x)
			}
		}
		slices.Sort(rest)
		return nil, &CycleError{Nodes: rest}
	}
	return order, nil
}
