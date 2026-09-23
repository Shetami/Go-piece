package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoEquals = errors.New("нет знака =")
	ErrEmptyKey = errors.New("пустой ключ")
)

// ParseError сообщает, в какой строке случилась ошибка.
type ParseError struct {
	Line int
	Err  error
}

func (e *ParseError) Error() string { return fmt.Sprintf("строка %d: %v", e.Line, e.Err) }

// Unwrap позволяет проверять причину через errors.Is(err, ErrNoEquals).
func (e *ParseError) Unwrap() error { return e.Err }

// ParseConfig разбирает строки вида "key=value". Пустые строки и строки,
// начинающиеся с #, пропускаются, пробелы вокруг ключа и значения срезаются.
// При ошибке возвращает *ParseError с номером строки (с единицы).
func ParseConfig(text string) (map[string]string, error) {
	cfg := make(map[string]string)
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, &ParseError{Line: i + 1, Err: ErrNoEquals}
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, &ParseError{Line: i + 1, Err: ErrEmptyKey}
		}
		cfg[key] = strings.TrimSpace(value)
	}
	return cfg, nil
}
