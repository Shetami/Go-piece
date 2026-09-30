package main

import (
	"errors"
	"fmt"
	"io"
)

// Directive — одна директива конфига: "listen 0.0.0.0 8080" →
// {Line: N, Key: "listen", Args: ["0.0.0.0", "8080"]}.
type Directive struct {
	Line int // номер строки (с 1), где директива начинается
	Key  string
	Args []string
}

// LineError — ошибка в конкретной строке конфига.
type LineError struct {
	Line int
	Err  error
}

func (e *LineError) Error() string { return fmt.Sprintf("строка %d: %v", e.Line, e.Err) }
func (e *LineError) Unwrap() error { return e.Err }

// ErrUnclosedQuote — в директиве есть незакрытая кавычка.
var ErrUnclosedQuote = errors.New("незакрытая кавычка")

// ParseConfig читает конфиг из r построчно.
//
//   - окончания строк — "\n" или "\r\n"; последняя строка может быть без "\n";
//   - строка, которая заканчивается на '\', продолжается на следующей:
//     '\' убирается, строки склеиваются через пробел; Line директивы —
//     номер её первой строки;
//   - '#' вне кавычек начинает комментарий до конца строки;
//   - слова разделены пробелами и табами; "..." — одно слово, в котором
//     могут быть пробелы и '#', сами кавычки в слово не входят;
//   - первое слово — Key, остальные — Args;
//   - пустые строки и строки только с комментарием пропускаются;
//   - строки бывают очень длинными (сотни килобайт) — это не ошибка.
//
// Незакрытая кавычка — *LineError{Line, ErrUnclosedQuote} с номером первой
// строки директивы. Ошибка чтения из r возвращается обёрнутой (errors.Is
// её находит). При ошибке результат — nil.
func ParseConfig(r io.Reader) ([]Directive, error) {
	// ваш код
	return nil, nil
}
