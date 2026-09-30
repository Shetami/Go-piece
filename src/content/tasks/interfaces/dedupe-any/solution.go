package main

import (
	"math"
	"reflect"
)

// Keyer — значение само сообщает ключ для сравнения.
type Keyer interface {
	DedupKey() string
}

type keyerKey struct {
	t   reflect.Type
	key string
}

// nanKey — отдельный ключ для NaN каждого типа: сам NaN как ключ мапы
// никогда не находится (NaN != NaN).
type nanKey struct{ t reflect.Type }

// Dedupe возвращает значения без повторов, сохраняя порядок первых вхождений.
// Одинаковыми считаются:
//   - значения, реализующие Keyer, одного динамического типа с одинаковым DedupKey();
//   - сравнимые значения, равные по == (int(1) и int64(1) — разные: разные типы);
//   - float32/float64 NaN — равен NaN того же типа;
//   - несравнимые значения (слайсы, мапы, структуры с ними, а также
//     структуры и массивы, в чьих полях-интерфейсах лежит несравнимое) —
//     если reflect.DeepEqual.
//
// nil и типизированный nil-указатель — разные значения. Функция не паникует
// ни на каких входах и не меняет xs.
func Dedupe(xs []any) []any {
	out := make([]any, 0, len(xs))
	seen := map[any]struct{}{}
	var loose []any // несравнимые: сравниваем попарно через DeepEqual
	add := func(k any, x any) {
		if _, dup := seen[k]; !dup {
			seen[k] = struct{}{}
			out = append(out, x)
		}
	}
	for _, x := range xs {
		if k, ok := x.(Keyer); ok {
			add(keyerKey{reflect.TypeOf(x), k.DedupKey()}, x)
			continue
		}
		if x == nil {
			add(nil, x) // nil-интерфейс — законный ключ мапы
			continue
		}
		switch f := x.(type) {
		case float64:
			if math.IsNaN(f) {
				add(nanKey{reflect.TypeOf(x)}, x)
				continue
			}
		case float32:
			if f != f {
				add(nanKey{reflect.TypeOf(x)}, x)
				continue
			}
		}
		// Value.Comparable (Go 1.20) проверяет и содержимое полей-интерфейсов:
		// struct{ A any }{[]int{1}} — тип сравнимый, а == на нём паникует.
		if reflect.ValueOf(x).Comparable() {
			add(x, x)
			continue
		}
		dup := false
		for _, y := range loose {
			if reflect.DeepEqual(x, y) {
				dup = true
				break
			}
		}
		if !dup {
			loose = append(loose, x)
			out = append(out, x)
		}
	}
	return out
}
