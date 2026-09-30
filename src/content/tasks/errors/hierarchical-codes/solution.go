package main

import (
	"errors"
	"strings"
)

// Code — иерархический код ошибки из сегментов через точку:
// "db", "db.conn", "db.conn.timeout". Сам по себе Code — тоже ошибка,
// поэтому коды служат сентинелами: errors.Is(err, Code("db")).
type Code string

func (c Code) Error() string { return string(c) }

// E — ошибка с кодом, сообщением и (необязательной) причиной.
type E struct {
	Code Code
	Msg  string
	Err  error
}

// Is: target — Code, и он совпадает с c или является его предком по
// целым сегментам. "db" — предок "db.conn.timeout"; "d", "db.co" и
// "db.conn.timeout.x" — нет. Пустой код не совпадает ни с чем.
func (c Code) Is(target error) bool {
	t, ok := target.(Code)
	if !ok || t == "" {
		return false
	}
	// Префикс по сегментам: сравнить строку и проверить, что дальше точка.
	return c == t || strings.HasPrefix(string(c), string(t)+".")
}

// Error: "<code>: <msg>", а если есть причина — ещё ": <причина>".
func (e *E) Error() string {
	s := string(e.Code) + ": " + e.Msg
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

// Unwrap открывает причину.
func (e *E) Unwrap() error {
	return e.Err
}

// Is: errors.Is(e, Code(x)) — по тем же правилам иерархии, что и Code.Is.
// По причине errors.Is пройдёт сам — через Unwrap.
func (e *E) Is(target error) bool {
	return e.Code.Is(target)
}

// CodeOf возвращает код самого внешнего *E в цепочке или "", если его нет.
func CodeOf(err error) Code {
	var e *E
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}
