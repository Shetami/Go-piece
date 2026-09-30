package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrDuplicateKey — ключ встретился второй раз.
var ErrDuplicateKey = errors.New("повторный ключ")

// ParseError — ошибка разбора. Pos — позиция в символах (рунах, с нуля),
// на которой разбор споткнулся. Err — причина: ErrDuplicateKey или nil.
type ParseError struct {
	Pos int
	Msg string
	Err error
}

func (e *ParseError) Error() string { return fmt.Sprintf("позиция %d: %s", e.Pos, e.Msg) }
func (e *ParseError) Unwrap() error { return e.Err }

// ParseKV разбирает строку вида
//
//	host=db.local port=5432 name="my app" note="say \"hi\"" empty=
//
// Правила:
//   - пары разделены одним или несколькими пробелами, пробелы по краям допустимы;
//   - ключ непустой, из букв (любых, в том числе кириллицы), цифр и символов
//     '_', '-', '.'; сразу за ключом — '=';
//   - значение без кавычек идёт до пробела или конца строки, может быть
//     пустым и не может содержать '"';
//   - значение в кавычках может содержать пробелы и '='; внутри работают
//     экранирования \" и \\, любое другое \x — ошибка; за закрывающей
//     кавычкой — пробел или конец строки.
//
// Любая ошибка — *ParseError (и nil вместо мапы):
//   - недопустимый символ в ключе или нет '=' — Pos этого символа
//     (или длина строки, если она кончилась);
//   - пустой ключ — Pos знака '=';
//   - '"' внутри значения без кавычек — Pos этой кавычки;
//   - незакрытая кавычка — Pos открывающей кавычки;
//   - неизвестное экранирование — Pos обратной косой черты;
//   - мусор сразу за закрывающей кавычкой — Pos этого символа;
//   - повторный ключ — Pos начала повторного ключа, Err = ErrDuplicateKey;
//     если в повторной паре есть синтаксическая ошибка, возвращается она.
func ParseKV(s string) (map[string]string, error) {
	rs := []rune(s) // позиции — в символах, поэтому работаем со слайсом рун
	n := len(rs)
	res := map[string]string{}
	fail := func(pos int, msg string) (map[string]string, error) {
		return nil, &ParseError{Pos: pos, Msg: msg}
	}
	i := 0
	for {
		for i < n && rs[i] == ' ' {
			i++
		}
		if i == n {
			return res, nil
		}
		// Ключ.
		start := i
		for i < n && isKeyRune(rs[i]) {
			i++
		}
		if i == n || rs[i] != '=' {
			return fail(i, "ожидали '='")
		}
		if i == start {
			return fail(i, "пустой ключ")
		}
		key := string(rs[start:i])
		i++ // '='

		// Значение.
		var val strings.Builder
		if i < n && rs[i] == '"' {
			open := i
			i++
			closed := false
			for i < n && !closed {
				switch r := rs[i]; r {
				case '"':
					closed = true
				case '\\':
					if i+1 < n && (rs[i+1] == '"' || rs[i+1] == '\\') {
						val.WriteRune(rs[i+1])
						i++
					} else {
						return fail(i, "неизвестное экранирование")
					}
				default:
					val.WriteRune(r)
				}
				i++
			}
			if !closed {
				return fail(open, "незакрытая кавычка")
			}
			if i < n && rs[i] != ' ' {
				return fail(i, "ожидали пробел после кавычки")
			}
		} else {
			for i < n && rs[i] != ' ' {
				if rs[i] == '"' {
					return fail(i, "кавычка внутри значения")
				}
				val.WriteRune(rs[i])
				i++
			}
		}
		// Дубликат проверяем после разбора значения: синтаксическая
		// ошибка в этой же паре важнее.
		if _, dup := res[key]; dup {
			return nil, &ParseError{Pos: start, Msg: "повторный ключ " + key, Err: ErrDuplicateKey}
		}
		res[key] = val.String()
	}
}

func isKeyRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.'
}
