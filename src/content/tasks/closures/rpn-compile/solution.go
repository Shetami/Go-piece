package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrSyntax  = errors.New("syntax error")
	ErrUnknown = errors.New("unknown variable")
	ErrDivZero = errors.New("division by zero")
)

// Expr вычисляет скомпилированное выражение на наборе переменных.
type Expr func(vars map[string]float64) (float64, error)

// Compile разбирает выражение в обратной польской записи: токены через
// пробелы — числа (как в strconv.ParseFloat), имена переменных (буквы,
// цифры, _; начинаются не с цифры) и операторы + - * /.
// "a b -" — это a - b.
//
// Весь разбор и все проверки структуры — в Compile: пустое выражение,
// нехватка операндов, лишние операнды, неизвестный токен → ошибка с
// ErrSyntax. Возвращённая Expr только вычисляет: ей нельзя заново резать
// строку. Ошибки вычисления: переменной нет в vars → ErrUnknown (в тексте
// имя), деление на ноль → ErrDivZero.
func Compile(src string) (Expr, error) {
	var stack []Expr
	for _, tok := range strings.Fields(src) {
		switch tok {
		case "+", "-", "*", "/":
			if len(stack) < 2 {
				return nil, fmt.Errorf("%w: %q needs two operands", ErrSyntax, tok)
			}
			// Правый операнд — верхний: "a b -" это a - b.
			l, r := stack[len(stack)-2], stack[len(stack)-1]
			stack = append(stack[:len(stack)-2], binary(tok, l, r))
			continue
		}
		if n, err := strconv.ParseFloat(tok, 64); err == nil {
			stack = append(stack, func(map[string]float64) (float64, error) { return n, nil })
			continue
		}
		if !isIdent(tok) {
			return nil, fmt.Errorf("%w: bad token %q", ErrSyntax, tok)
		}
		name := tok
		stack = append(stack, func(vars map[string]float64) (float64, error) {
			v, ok := vars[name]
			if !ok {
				return 0, fmt.Errorf("%w: %s", ErrUnknown, name)
			}
			return v, nil
		})
	}
	if len(stack) != 1 {
		return nil, fmt.Errorf("%w: %d values left on stack", ErrSyntax, len(stack))
	}
	return stack[0], nil
}

// binary собирает узел-замыкание: оператор выбран один раз, при компиляции.
func binary(op string, l, r Expr) Expr {
	var f func(a, b float64) (float64, error)
	switch op {
	case "+":
		f = func(a, b float64) (float64, error) { return a + b, nil }
	case "-":
		f = func(a, b float64) (float64, error) { return a - b, nil }
	case "*":
		f = func(a, b float64) (float64, error) { return a * b, nil }
	case "/":
		f = func(a, b float64) (float64, error) {
			if b == 0 {
				return 0, ErrDivZero
			}
			return a / b, nil
		}
	}
	return func(vars map[string]float64) (float64, error) {
		a, err := l(vars)
		if err != nil {
			return 0, err
		}
		b, err := r(vars)
		if err != nil {
			return 0, err
		}
		return f(a, b)
	}
}

func isIdent(s string) bool {
	for i, c := range s {
		letter := c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		digit := c >= '0' && c <= '9'
		if !letter && !(digit && i > 0) {
			return false
		}
	}
	return s != ""
}
