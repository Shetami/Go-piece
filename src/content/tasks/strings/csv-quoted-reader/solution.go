package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	// ErrQuote — кавычка не на своём месте или не закрыта.
	ErrQuote = errors.New("ошибка кавычек")
	// ErrFieldCount — в записи не столько полей, сколько в первой.
	ErrFieldCount = errors.New("неверное число полей")
)

// CSVError — ошибка в записи, которая начинается на строке Line (с 1).
type CSVError struct {
	Line int
	Err  error
}

func (e *CSVError) Error() string { return fmt.Sprintf("csv, строка %d: %v", e.Line, e.Err) }
func (e *CSVError) Unwrap() error { return e.Err }

// ReadCSV читает все записи CSV из r. Разделитель — запятая.
// Пакет encoding/csv не используется — это и есть задача.
//
//   - записи разделены "\n" или "\r\n", последняя может быть без перевода строки;
//   - пустые строки (вне кавычек) пропускаются и записей не дают;
//   - поле в кавычках целиком: "a,b" — одно поле; внутри могут быть запятые,
//     переводы строк и удвоенная кавычка "" (означает одну "); "\r\n" внутри
//     кавычек превращается в "\n";
//   - "a,b," — три поля, последнее пустое; "" — пустое поле;
//   - кавычка внутри поля без кавычек (a"b) или что-то кроме запятой и конца
//     строки сразу после закрывающей кавычки ("a"b) — ErrQuote;
//   - незакрытая кавычка — тоже ErrQuote;
//   - у всех записей столько же полей, сколько у первой, иначе ErrFieldCount.
//
// Ошибки формата — *CSVError, Line — строка, на которой началась запись
// (строки считаются по "\n", включая переводы строк внутри кавычек). Ошибка чтения из r возвращается обёрнутой.
// При ошибке результат — nil.
func ReadCSV(r io.Reader) ([][]string, error) {
	br := bufio.NewReader(r)
	var records [][]string
	line := 1
	for {
		// Пропускаем пустые строки между записями.
		c, _, err := br.ReadRune()
		for err == nil && (c == '\n' || c == '\r') {
			if c == '\n' {
				line++
			}
			c, _, err = br.ReadRune()
		}
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, fmt.Errorf("чтение csv: %w", err)
		}
		if uerr := br.UnreadRune(); uerr != nil {
			return nil, uerr
		}
		start := line
		rec, err := readRecord(br, &line)
		if err != nil {
			var ce *CSVError
			if errors.As(err, &ce) {
				ce.Line = start
			}
			return nil, err
		}
		if len(records) > 0 && len(rec) != len(records[0]) {
			return nil, &CSVError{Line: start, Err: ErrFieldCount}
		}
		records = append(records, rec)
	}
}

// readRecord читает одну запись; line увеличивается на каждый прочитанный '\n'.
func readRecord(br *bufio.Reader, line *int) ([]string, error) {
	var rec []string
	var f strings.Builder
	for {
		quoted := false
		c, _, err := br.ReadRune()
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("чтение csv: %w", err)
		}
		// Поле в кавычках.
		if err == nil && c == '"' {
			quoted = true
			for {
				c, _, err = br.ReadRune()
				if err == io.EOF {
					return nil, &CSVError{Err: fmt.Errorf("%w: незакрытая кавычка", ErrQuote)}
				}
				if err != nil {
					return nil, fmt.Errorf("чтение csv: %w", err)
				}
				if c == '"' {
					next, _, err2 := br.ReadRune()
					if err2 == nil && next == '"' {
						f.WriteRune('"') // "" внутри кавычек — одна кавычка
						continue
					}
					if err2 == nil {
						_ = br.UnreadRune()
					} else if err2 != io.EOF {
						return nil, fmt.Errorf("чтение csv: %w", err2)
					}
					break
				}
				if c == '\r' {
					// "\r\n" внутри кавычек сохраняем как "\n".
					if next, _, err2 := br.ReadRune(); err2 == nil && next == '\n' {
						c = '\n'
					} else if err2 == nil {
						_ = br.UnreadRune()
					}
				}
				if c == '\n' {
					*line++
				}
				f.WriteRune(c)
			}
			c, _, err = br.ReadRune()
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("чтение csv: %w", err)
			}
			if err == nil && c == '\r' {
				c, _, err = br.ReadRune()
				if err == nil && c != '\n' {
					return nil, &CSVError{Err: fmt.Errorf("%w: мусор после кавычки", ErrQuote)}
				}
			}
			if err == nil && c != ',' && c != '\n' {
				return nil, &CSVError{Err: fmt.Errorf("%w: мусор после кавычки", ErrQuote)}
			}
		} else {
			// Поле без кавычек — до запятой или конца строки.
			for err == nil && c != ',' && c != '\n' {
				if c == '"' {
					return nil, &CSVError{Err: fmt.Errorf("%w: кавычка внутри поля", ErrQuote)}
				}
				f.WriteRune(c)
				c, _, err = br.ReadRune()
			}
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("чтение csv: %w", err)
			}
		}
		val := f.String()
		f.Reset()
		rec = append(rec, val)
		if err == io.EOF || c == '\n' {
			if c == '\n' {
				*line++
			}
			if !quoted {
				// "\r\n": '\r' попал в конец поля без кавычек — отрезаем.
				rec[len(rec)-1] = strings.TrimSuffix(val, "\r")
			}
			return rec, nil
		}
		// c == ',' — следующее поле.
	}
}
