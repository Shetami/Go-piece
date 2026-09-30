package main

import (
	"cmp"
	"slices"
)

// Row — строка таблицы с уникальным ID.
type Row struct {
	ID  int
	Val string
}

// Upsert вливает batch в rows и возвращает результат, отсортированный по ID.
// rows отсортированы по ID, ID уникальны. batch — в любом порядке, ID в нём
// могут повторяться: побеждает последнее вхождение. Строка с существующим
// ID заменяется, новая — встаёт на своё место.
//
// rows можно менять на месте, и если его вместимости хватает, результат
// использует тот же массив. batch не меняется.
// Время — O(n + m log m), где n = len(rows), m = len(batch): не вставляйте
// строки по одной со сдвигом хвоста.
func Upsert(rows, batch []Row) []Row {
	b := slices.Clone(batch)
	// Стабильная сортировка сохраняет порядок повторов: последний в серии — последний во входе.
	slices.SortStableFunc(b, func(x, y Row) int { return cmp.Compare(x.ID, y.ID) })
	w := 0
	for i := range b {
		if i+1 < len(b) && b[i+1].ID == b[i].ID {
			continue // есть более позднее вхождение того же ID
		}
		b[w] = b[i]
		w++
	}
	b = b[:w]

	// Два указателя: существующие ID обновляем на месте, новые собираем в начало b.
	fresh := 0
	i := 0
	for _, r := range b {
		for i < len(rows) && rows[i].ID < r.ID {
			i++
		}
		if i < len(rows) && rows[i].ID == r.ID {
			rows[i] = r
		} else {
			b[fresh] = r
			fresh++
		}
	}
	b = b[:fresh]

	// Слияние с конца: пишем в хвост выросшего rows и не затираем непрочитанное.
	n := len(rows)
	rows = slices.Grow(rows, fresh)[:n+fresh] // не больше одной аллокации
	i, j := n-1, fresh-1
	for k := n + fresh - 1; j >= 0; k-- {
		if i >= 0 && rows[i].ID > b[j].ID {
			rows[k] = rows[i]
			i--
		} else {
			rows[k] = b[j]
			j--
		}
	}
	return rows
}
