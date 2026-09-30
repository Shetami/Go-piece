package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	ErrConflict = errors.New("путь одновременно лист и узел")
	ErrBadPath  = errors.New("пустой сегмент пути")
)

// Unflatten собирает вложенный документ из плоской мапы «путь → значение»,
// сегменты пути разделены точкой:
//
//	{"db.host": "x", "db.port": 5432, "debug": true}
//	→ {"db": {"host": "x", "port": 5432}, "debug": true}
//
//   - Все промежуточные узлы — map[string]any ("l.0" даёт ключ "0", слайсы не
//     восстанавливаются).
//   - Значение-лист кладётся как есть, даже если оно само map[string]any, и
//     внутрь него ничего не дописывается.
//   - Если путь — одновременно лист и префикс другого пути ("a": 1 и "a.b": 2),
//     ошибка errors.Is(err, ErrConflict) — при любом порядке обхода мапы.
//   - Путь с пустым сегментом ("", "a..b", ".a", "a.") — ErrBadPath.
//   - При ошибке результат nil. Вход не меняется.
func Unflatten(flat map[string]any) (map[string]any, error) {
	// Порядок обхода мапы случаен — сортируем пути, чтобы ответ не зависел от него.
	paths := make([]string, 0, len(flat))
	for p := range flat {
		paths = append(paths, p)
	}
	slices.Sort(paths)

	root := make(map[string]any)
	// Узлы, которые создали мы сами. Только в них можно спускаться:
	// map[string]any, пришедшая как значение, — это лист.
	ours := map[string]bool{"": true}

	for _, p := range paths {
		segs := strings.Split(p, ".")
		if slices.Contains(segs, "") {
			return nil, fmt.Errorf("%w: %q", ErrBadPath, p)
		}
		node, prefix := root, ""
		for _, s := range segs[:len(segs)-1] {
			if prefix == "" {
				prefix = s
			} else {
				prefix += "." + s
			}
			next, exists := node[s]
			if !exists {
				m := make(map[string]any)
				node[s] = m
				ours[prefix] = true
				node = m
				continue
			}
			if !ours[prefix] {
				return nil, fmt.Errorf("%w: %q и %q", ErrConflict, prefix, p)
			}
			node = next.(map[string]any)
		}
		last := segs[len(segs)-1]
		if _, exists := node[last]; exists {
			// Здесь уже узел, созданный для более длинного пути.
			return nil, fmt.Errorf("%w: %q", ErrConflict, p)
		}
		node[last] = flat[p]
	}
	return root, nil
}
