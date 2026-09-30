package main

import (
	"errors"
	"fmt"
)

// ParseError — ошибка вычисления с позицией (смещение в байтах от 0).
type ParseError struct {
	Pos int
	Msg string
	Err error // причина, если есть (например, ErrDivByZero)
}

func (e *ParseError) Error() string { return fmt.Sprintf("позиция %d: %s", e.Pos, e.Msg) }
func (e *ParseError) Unwrap() error { return e.Err }

var ErrDivByZero = errors.New("деление на ноль")

// Eval вычисляет целочисленное выражение: числа, идентификаторы
// (латинские буквы и _), + - * / (деление целое), скобки, унарный минус.
// Пробелы игнорируются. Приоритеты обычные, операторы левоассоциативны.
// Значение идентификатора возвращает lookup.
//
// Ошибки — всегда *ParseError:
//   - синтаксис: Pos — позиция первого неожиданного символа,
//     или len(expr), если выражение неожиданно кончилось;
//   - деление на ноль: Pos — позиция знака '/', Err = ErrDivByZero.
//
// Паника внутри lookup — чужая: любая (даже значением *ParseError)
// летит из Eval наружу как есть.
func Eval(expr string, lookup func(name string) int) (int, error) {
	// ваш код
	return 0, nil
}
