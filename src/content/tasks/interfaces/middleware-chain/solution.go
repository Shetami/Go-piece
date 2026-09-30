package main

import (
	"context"
	"errors"
	"fmt"
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
	return f(ctx, r)
}

// Middleware оборачивает обработчик.
type Middleware func(next Handler) Handler

// Chain собирает цепочку: Chain(h, a, b) — запрос проходит a → b → h,
// ответ возвращается в обратном порядке. Без middleware — это сам h.
func Chain(h Handler, mws ...Middleware) Handler {
	// Оборачиваем с конца: последний middleware — ближе всех к h.
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ErrPanic — обработчик паниковал.
var ErrPanic = errors.New("handler panicked")

// Recover превращает панику в next в ошибку: ответ nil, errors.Is(err, ErrPanic),
// текст содержит значение паники; если значение паники — error, то
// errors.Is работает и для него. Обычные ошибки проходят как есть.
func Recover() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, r *Request) (resp *Response, err error) {
			defer func() {
				if v := recover(); v != nil {
					resp = nil
					if e, ok := v.(error); ok {
						err = fmt.Errorf("%w: %w", ErrPanic, e) // обе ошибки в цепочке
					} else {
						err = fmt.Errorf("%w: %v", ErrPanic, v)
					}
				}
			}()
			return next.Serve(ctx, r)
		})
	}
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
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, r *Request) (*Response, error) {
			resp, err := next.Serve(ctx, r)
			if err != nil {
				status := 500
				var se StatusError
				if errors.As(err, &se) { // As с указателем на интерфейс ищет любой тип, реализующий его
					status = se.Status()
				}
				return &Response{Status: status, Body: err.Error()}, nil
			}
			if resp == nil {
				return &Response{Status: 500, Body: "empty response"}, nil
			}
			return resp, nil
		})
	}
}

type statusErr struct {
	code int
	msg  string
}

func (e statusErr) Error() string { return e.msg }
func (e statusErr) Status() int   { return e.code }

// RequireUser: пустой r.User — ошибка-StatusError со статусом 401 и
// текстом "unauthorized"; next при этом не вызывается.
func RequireUser() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, r *Request) (*Response, error) {
			if r.User == "" {
				return nil, statusErr{401, "unauthorized"}
			}
			return next.Serve(ctx, r)
		})
	}
}
