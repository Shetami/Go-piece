package main

import (
	"encoding"
	"errors"
	"fmt"
	"strconv"
	"time"
)

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
	s, ok := env[key]
	if !ok {
		return def, nil
	}
	var v T
	var err error
	// Переключаемся по типу указателя на v: так можно и узнать T, и
	// записать в v, не приводя результат обратно к T.
	switch p := any(&v).(type) {
	case *string:
		*p = s
	case *int:
		*p, err = strconv.Atoi(s)
	case *int64:
		*p, err = strconv.ParseInt(s, 10, 64)
	case *float64:
		*p, err = strconv.ParseFloat(s, 64)
	case *bool:
		*p, err = strconv.ParseBool(s)
	case *time.Duration: // отдельный тип, хоть в основе и int64
		*p, err = time.ParseDuration(s)
	case encoding.TextUnmarshaler:
		err = p.UnmarshalText([]byte(s))
	default:
		return def, fmt.Errorf("%s: %T: %w", key, v, ErrUnsupported)
	}
	if err != nil {
		return def, fmt.Errorf("%s: %w", key, err)
	}
	return v, nil
}
