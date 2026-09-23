package main

import "context"

// WithRequestID кладёт идентификатор запроса в контекст.
// Прочитать его можно только через RequestID — никакой другой код,
// даже знающий строку "request_id", до него не доберётся.
func WithRequestID(ctx context.Context, id string) context.Context {
	// ваш код
	return ctx
}

// RequestID достаёт идентификатор запроса. ok == false, если его нет.
func RequestID(ctx context.Context) (string, bool) {
	// ваш код
	return "", false
}
