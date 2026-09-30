package main

import (
	"context"
	"errors"
)

type Request struct {
	Path string
	User string
}

type Response struct {
	Status int
	Body   string
}

// Handler обрабатывает запрос.
type Handler interface {
	Serve(ctx context.Context, r *Request) (*Response, error)
}

// HandlerFunc превращает обычную функцию в Handler.
type HandlerFunc func(ctx context.Context, r *Request) (*Response, error)

// Serve вызывает f.
func (f HandlerFunc) Serve(ctx context.Context, r *Request) (*Response, error) {
	// ваш код
	return nil, nil
}

// Middleware оборачивает обработчик.
type Middleware func(next Handler) Handler

// Chain собирает цепочку: Chain(h, a, b) — запрос проходит a → b → h,
// ответ возвращается в обратном порядке. Без middleware — это сам h.
func Chain(h Handler, mws ...Middleware) Handler {
	// ваш код
	return h
}

// ErrPanic — обработчик паниковал.
var ErrPanic = errors.New("handler panicked")

// Recover превращает панику в next в ошибку: ответ nil, errors.Is(err, ErrPanic),
// текст содержит значение паники; если значение паники — error, то
// errors.Is работает и для него. Обычные ошибки проходят как есть.
func Recover() Middleware {
	// ваш код
	return func(next Handler) Handler { return next }
}

// StatusError — ошибка, знающая свой HTTP-статус.
type StatusError interface {
	error
	Status() int
}

// Errors превращает ошибки next в ответы и всегда возвращает nil-ошибку:
//   - в цепочке обёрток есть StatusError — его статус; иначе 500;
//   - Body — полный текст ошибки err.Error();
//   - next вернул (nil, nil) — ответ {500, "empty response"};
//   - успешный ответ проходит как есть.
func Errors() Middleware {
	// ваш код
	return func(next Handler) Handler { return next }
}

// RequireUser: пустой r.User — ошибка-StatusError со статусом 401 и
// текстом "unauthorized"; next при этом не вызывается.
func RequireUser() Middleware {
	// ваш код
	return func(next Handler) Handler { return next }
}
