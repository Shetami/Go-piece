package main

import (
	"context"
	"fmt"
	"strings"
)

type Request struct {
	Header map[string]string // может быть nil
	Body   string
}

type Handler func(ctx context.Context, req Request) string
type Middleware func(Handler) Handler

// Приватные типы ключей: никто вне пакета не прочитает и не подменит.
type ridKey struct{}
type userKey struct{}

// Chain оборачивает h в middleware; mws[0] — самый внешний (выполняется первым).
func Chain(h Handler, mws ...Middleware) Handler {
	for i := len(mws) - 1; i >= 0; i-- { // оборачиваем с конца
		h = mws[i](h)
	}
	return h
}

// WithRequestID кладёт в контекст request id: из заголовка X-Request-Id
// (без пробелов по краям), а если его нет или он пустой — gen().
func WithRequestID(gen func() string) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req Request) string {
			id := strings.TrimSpace(req.Header["X-Request-Id"]) // чтение из nil-мапы безопасно
			if id == "" {
				id = gen()
			}
			return next(context.WithValue(ctx, ridKey{}, id), req)
		}
	}
}

// WithUser кладёт в контекст пользователя по заголовку
// "Authorization: Bearer <token>" и таблице tokens. Нет заголовка, другая
// схема или неизвестный токен — пользователя в контексте нет (а если его
// положил кто-то раньше — он не должен «просочиться» в этот запрос).
func WithUser(tokens map[string]string) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req Request) string {
			var user string
			tok, ok := strings.CutPrefix(req.Header["Authorization"], "Bearer ")
			if ok && tok != "" {
				user = tokens[tok]
			}
			// Кладём всегда: "" перекрывает чужого пользователя выше по цепочке.
			return next(context.WithValue(ctx, userKey{}, user), req)
		}
	}
}

// RequestID — request id из контекста или "".
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(ridKey{}).(string)
	return id
}

// User — пользователь из контекста; ok == false, если его нет.
func User(ctx context.Context) (string, bool) {
	u, _ := ctx.Value(userKey{}).(string)
	return u, u != ""
}

// Logf форматирует строку лога с префиксом "[<rid> <user>] ".
// Нет request id — "-", нет пользователя — "anon".
func Logf(ctx context.Context, format string, args ...any) string {
	rid := RequestID(ctx)
	if rid == "" {
		rid = "-"
	}
	user, ok := User(ctx)
	if !ok {
		user = "anon"
	}
	return fmt.Sprintf("[%s %s] ", rid, user) + fmt.Sprintf(format, args...)
}
