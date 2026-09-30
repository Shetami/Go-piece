package main

import "errors"

// Table хранит данные по колонкам: строка i — это Names[i], Ages[i], Scores[i].
type Table struct {
	Names  []string
	Ages   []int
	Scores []float64
}

var ErrRagged = errors.New("columns have different lengths")

// SortBy сортирует строки таблицы по ключам "name", "age", "score";
// префикс "-" — по убыванию. Правила:
//   - сортировка стабильная: строки, равные по всем ключам, сохраняют порядок;
//   - NaN в Scores — всегда в конце, и при возрастании, и при убывании;
//   - пустой список ключей или неизвестный ключ — ошибка, таблица не меняется;
//   - колонки разной длины — ErrRagged, таблица не меняется.
//
// Решение — через sort.Interface и sort.Stable.
func SortBy(t *Table, keys ...string) error {
	// ваш код
	return nil
}
