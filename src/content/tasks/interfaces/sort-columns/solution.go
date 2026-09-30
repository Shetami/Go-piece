package main

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

// Table хранит данные по колонкам: строка i — это Names[i], Ages[i], Scores[i].
type Table struct {
	Names  []string
	Ages   []int
	Scores []float64
}

var ErrRagged = errors.New("columns have different lengths")

type sortKey struct {
	col  string
	desc bool
}

// byKeys реализует sort.Interface поверх всех колонок сразу.
type byKeys struct {
	t    *Table
	keys []sortKey
}

func (b byKeys) Len() int { return len(b.t.Names) }

// Swap обязан менять строку целиком — все колонки.
func (b byKeys) Swap(i, j int) {
	t := b.t
	t.Names[i], t.Names[j] = t.Names[j], t.Names[i]
	t.Ages[i], t.Ages[j] = t.Ages[j], t.Ages[i]
	t.Scores[i], t.Scores[j] = t.Scores[j], t.Scores[i]
}

func (b byKeys) Less(i, j int) bool {
	for _, k := range b.keys {
		var c int
		switch k.col {
		case "name":
			c = strings.Compare(b.t.Names[i], b.t.Names[j])
		case "age":
			c = cmp.Compare(b.t.Ages[i], b.t.Ages[j])
		case "score":
			x, y := b.t.Scores[i], b.t.Scores[j]
			xn, yn := math.IsNaN(x), math.IsNaN(y)
			if xn || yn {
				// NaN всегда в конце — независимо от направления, поэтому до инверсии.
				if xn && yn {
					continue
				}
				return yn
			}
			c = cmp.Compare(x, y)
		}
		if k.desc {
			c = -c // инвертируем сравнение, а не результат Less: равные остаются равными
		}
		if c != 0 {
			return c < 0
		}
	}
	return false
}

// SortBy сортирует строки таблицы по ключам "name", "age", "score";
// префикс "-" — по убыванию. Правила:
//   - сортировка стабильная: строки, равные по всем ключам, сохраняют порядок;
//   - NaN в Scores — всегда в конце, и при возрастании, и при убывании;
//   - пустой список ключей или неизвестный ключ — ошибка, таблица не меняется;
//   - колонки разной длины — ErrRagged, таблица не меняется.
//
// Решение — через sort.Interface и sort.Stable.
func SortBy(t *Table, keys ...string) error {
	if len(t.Ages) != len(t.Names) || len(t.Scores) != len(t.Names) {
		return ErrRagged
	}
	if len(keys) == 0 {
		return errors.New("no sort keys")
	}
	// Все ключи проверяем до начала сортировки: иначе таблица останется наполовину переставленной.
	parsed := make([]sortKey, 0, len(keys))
	for _, k := range keys {
		sk := sortKey{col: strings.TrimPrefix(k, "-"), desc: strings.HasPrefix(k, "-")}
		switch sk.col {
		case "name", "age", "score":
		default:
			return fmt.Errorf("unknown sort key %q", k)
		}
		parsed = append(parsed, sk)
	}
	sort.Stable(byKeys{t, parsed})
	return nil
}
