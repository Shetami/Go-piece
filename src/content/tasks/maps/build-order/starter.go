package main

import (
	"errors"
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
	// ваш код
	return nil, nil
}
