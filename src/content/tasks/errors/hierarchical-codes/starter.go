package main

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
	// ваш код
	return false
}

// Error: "<code>: <msg>", а если есть причина — ещё ": <причина>".
func (e *E) Error() string {
	// ваш код
	return ""
}

// Unwrap открывает причину.
func (e *E) Unwrap() error {
	// ваш код
	return nil
}

// Is: errors.Is(e, Code(x)) — по тем же правилам иерархии, что и Code.Is.
// По причине errors.Is пройдёт сам — через Unwrap.
func (e *E) Is(target error) bool {
	// ваш код
	return false
}

// CodeOf возвращает код самого внешнего *E в цепочке или "", если его нет.
func CodeOf(err error) Code {
	// ваш код
	return ""
}
