package main

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrKeyConflict = errors.New("два значения претендуют на один путь")

// Flatten разворачивает вложенный JSON-подобный документ в плоскую мапу
// «путь → значение», сегменты пути склеиваются точкой:
//
//	{"db": {"host": "x", "ports": [5432, 5433]}}
//	→ {"db.host": "x", "db.ports.0": 5432, "db.ports.1": 5433}
//
//   - Контейнеры — только map[string]any и []any; всё остальное (включая nil
//     и, например, map[string]int) — лист.
//   - Пустая map[string]any и пустой []any — тоже листья: они сохраняются под
//     своим путём как есть, иначе информация о них теряется.
//   - Если два разных места документа дают один путь ({"a.b": 1, "a": {"b": 2}}),
//     возвращается ошибка, для которой errors.Is(err, ErrKeyConflict), —
//     при любом порядке обхода мапы.
//   - Входной документ не меняется.
func Flatten(doc map[string]any) (map[string]any, error) {
	out := make(map[string]any)
	var walk func(path string, v any) error
	put := func(path string, v any) error {
		// Проверяем наличие ключа, а не значение: лист может быть nil.
		if _, dup := out[path]; dup {
			return fmt.Errorf("%w: %q", ErrKeyConflict, path)
		}
		out[path] = v
		return nil
	}
	join := func(path, seg string) string {
		if path == "" {
			return seg
		}
		return path + "." + seg
	}
	walk = func(path string, v any) error {
		switch x := v.(type) {
		case map[string]any:
			if len(x) == 0 {
				return put(path, x)
			}
			for k, child := range x {
				if err := walk(join(path, k), child); err != nil {
					return err
				}
			}
		case []any:
			if len(x) == 0 {
				return put(path, x)
			}
			for i, child := range x {
				if err := walk(join(path, strconv.Itoa(i)), child); err != nil {
					return err
				}
			}
		default:
			return put(path, x)
		}
		return nil
	}
	for k, v := range doc {
		if err := walk(k, v); err != nil {
			return nil, err
		}
	}
	return out, nil
}
