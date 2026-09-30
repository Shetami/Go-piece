package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
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
	// bufio.Reader, а не Scanner: у Scanner лимит строки 64 КБ по умолчанию.
	br := bufio.NewReader(r)
	var out []Directive
	var pending strings.Builder // склеенные строки-продолжения
	hasPending := false
	lineNo, startLine := 0, 0
	flush := func() error {
		if !hasPending {
			return nil
		}
		words, err := splitWords(pending.String())
		pending.Reset()
		hasPending = false
		if err != nil {
			return &LineError{Line: startLine, Err: err}
		}
		if len(words) > 0 {
			out = append(out, Directive{Line: startLine, Key: words[0], Args: words[1:]})
		}
		return nil
	}
	for {
		raw, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("чтение конфига: %w", err)
		}
		if raw != "" {
			lineNo++
			line := strings.TrimSuffix(strings.TrimSuffix(raw, "\n"), "\r")
			if !hasPending {
				startLine, hasPending = lineNo, true
			} else {
				pending.WriteByte(' ')
			}
			cont, isCont := strings.CutSuffix(line, `\`)
			if isCont {
				pending.WriteString(cont) // директива продолжится на следующей строке
			} else {
				pending.WriteString(line)
				if ferr := flush(); ferr != nil {
					return nil, ferr
				}
			}
		}
		if err == io.EOF {
			// Файл кончился сразу после '\' — дописываем то, что накопили.
			if ferr := flush(); ferr != nil {
				return nil, ferr
			}
			return out, nil
		}
	}
}

// splitWords режет строку на слова с учётом кавычек и комментариев.
func splitWords(s string) ([]string, error) {
	var words []string
	var cur strings.Builder
	inWord, inQuote := false, false
loop:
	for _, r := range s {
		switch {
		case inQuote:
			if r == '"' {
				inQuote = false
			} else {
				cur.WriteRune(r)
			}
		case r == '"':
			inQuote, inWord = true, true // "" — тоже слово, пустое
		case r == '#':
			break loop // комментарий только вне кавычек
		case r == ' ' || r == '\t':
			if inWord {
				words = append(words, cur.String())
				cur.Reset()
				inWord = false
			}
		default:
			cur.WriteRune(r)
			inWord = true
		}
	}
	if inQuote {
		return nil, ErrUnclosedQuote
	}
	if inWord {
		words = append(words, cur.String())
	}
	return words, nil
}
