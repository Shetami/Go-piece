package main

import "errors"

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
	// ваш код
	return nil, nil
}
