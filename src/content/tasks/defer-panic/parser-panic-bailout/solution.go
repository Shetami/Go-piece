package main

import (
	"errors"
	"fmt"
	"strconv"
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

// bailout — метка «это наша паника». Неэкспортируемый тип не может
// создать никто, кроме парсера, поэтому чужую панику с ним не спутать.
type bailout struct{ err *ParseError }

type parser struct {
	s      string
	pos    int
	lookup func(string) int
}

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
func Eval(expr string, lookup func(name string) int) (v int, err error) {
	defer func() {
		if r := recover(); r != nil {
			b, ok := r.(bailout)
			if !ok {
				panic(r) // не наша — отпускаем
			}
			v, err = 0, b.err
		}
	}()
	p := &parser{s: expr, lookup: lookup}
	v = p.expr()
	if p.peek() != 0 {
		p.fail(p.pos, fmt.Sprintf("лишний символ %q", p.s[p.pos]), nil)
	}
	return v, nil
}

// fail выходит из любой глубины рекурсии сразу в Eval.
func (p *parser) fail(pos int, msg string, err error) {
	panic(bailout{&ParseError{Pos: pos, Msg: msg, Err: err}})
}

// peek пропускает пробелы и возвращает текущий байт (0 — конец строки).
func (p *parser) peek() byte {
	for p.pos < len(p.s) && p.s[p.pos] == ' ' {
		p.pos++
	}
	if p.pos >= len(p.s) {
		return 0
	}
	return p.s[p.pos]
}

func (p *parser) expr() int {
	v := p.term()
	for {
		switch p.peek() {
		case '+':
			p.pos++
			v += p.term()
		case '-':
			p.pos++
			v -= p.term()
		default:
			return v
		}
	}
}

func (p *parser) term() int {
	v := p.unary()
	for {
		switch p.peek() {
		case '*':
			p.pos++
			v *= p.unary()
		case '/':
			opPos := p.pos
			p.pos++
			d := p.unary()
			if d == 0 {
				// Проверяем сами: иначе будет паника рантайма, а её
				// Eval честно отпустит наружу.
				p.fail(opPos, "деление на ноль", ErrDivByZero)
			}
			v /= d
		default:
			return v
		}
	}
}

func (p *parser) unary() int {
	if p.peek() == '-' {
		p.pos++
		return -p.unary()
	}
	return p.primary()
}

func isLetter(c byte) bool { return c == '_' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' }
func isDigit(c byte) bool  { return '0' <= c && c <= '9' }

func (p *parser) primary() int {
	c := p.peek()
	start := p.pos
	switch {
	case c == '(':
		p.pos++
		v := p.expr()
		if p.peek() != ')' {
			p.unexpected()
		}
		p.pos++
		return v
	case isDigit(c):
		for p.pos < len(p.s) && isDigit(p.s[p.pos]) {
			p.pos++
		}
		n, err := strconv.Atoi(p.s[start:p.pos])
		if err != nil {
			p.fail(start, "слишком большое число", err)
		}
		return n
	case isLetter(c):
		for p.pos < len(p.s) && (isLetter(p.s[p.pos]) || isDigit(p.s[p.pos])) {
			p.pos++
		}
		return p.lookup(p.s[start:p.pos])
	}
	p.unexpected()
	return 0
}

func (p *parser) unexpected() {
	if p.peek() == 0 {
		p.fail(p.pos, "неожиданный конец выражения", nil)
	}
	p.fail(p.pos, fmt.Sprintf("неожиданный символ %q", p.s[p.pos]), nil)
}
