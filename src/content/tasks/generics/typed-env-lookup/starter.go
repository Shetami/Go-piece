package main

import "errors"

// ErrUnsupported — Lookup не умеет разбирать такой тип.
var ErrUnsupported = errors.New("неподдерживаемый тип")

// Lookup читает переменную key из env и разбирает её в T.
//   - Ключа нет — (def, nil). Пустая строка — это значение, его разбираем.
//   - Поддерживаются ровно string, int, int64, float64, bool,
//     time.Duration (формат "1m30s") и любые T, у которых *T реализует
//     encoding.TextUnmarshaler. Свои типы вроде type Port int — нет.
//   - Ошибка разбора: (def, "<key>: <исходная ошибка>"), исходная
//     ошибка доступна через errors.Is/As.
//   - Неподдерживаемый тип: (def, ошибка c ErrUnsupported), без паники.
func Lookup[T any](env map[string]string, key string, def T) (T, error) {
	// ваш код
	return def, nil
}
