package main

import "errors"

var (
	ErrNoEquals = errors.New("нет знака =")
	ErrEmptyKey = errors.New("пустой ключ")
)

// ParseError сообщает, в какой строке случилась ошибка.
type ParseError struct {
	Line int
	Err  error
}

func (e *ParseError) Error() string { return "" }

// ParseConfig разбирает строки вида "key=value". Пустые строки и строки,
// начинающиеся с #, пропускаются, пробелы вокруг ключа и значения срезаются.
// При ошибке возвращает *ParseError с номером строки (с единицы).
func ParseConfig(text string) (map[string]string, error) {
	// ваш код
	return nil, nil
}
