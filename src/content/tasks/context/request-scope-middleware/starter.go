package main

import "context"

type Request struct {
	Header map[string]string // может быть nil
	Body   string
}

type Handler func(ctx context.Context, req Request) string
type Middleware func(Handler) Handler

// Chain оборачивает h в middleware; mws[0] — самый внешний (выполняется первым).
func Chain(h Handler, mws ...Middleware) Handler {
	// ваш код
	return h
}

// WithRequestID кладёт в контекст request id: из заголовка X-Request-Id
// (без пробелов по краям), а если его нет или он пустой — gen().
func WithRequestID(gen func() string) Middleware {
	// ваш код
	return func(next Handler) Handler { return next }
}

// WithUser кладёт в контекст пользователя по заголовку
// "Authorization: Bearer <token>" и таблице tokens. Нет заголовка, другая
// схема или неизвестный токен — пользователя в контексте нет (а если его
// положил кто-то раньше — он не должен «просочиться» в этот запрос).
func WithUser(tokens map[string]string) Middleware {
	// ваш код
	return func(next Handler) Handler { return next }
}

// RequestID — request id из контекста или "".
func RequestID(ctx context.Context) string {
	// ваш код
	return ""
}

// User — пользователь из контекста; ok == false, если его нет.
func User(ctx context.Context) (string, bool) {
	// ваш код
	return "", false
}

// Logf форматирует строку лога с префиксом "[<rid> <user>] ".
// Нет request id — "-", нет пользователя — "anon".
func Logf(ctx context.Context, format string, args ...any) string {
	// ваш код
	return ""
}
