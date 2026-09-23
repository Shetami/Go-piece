package main

import "context"

// requestIDKey — неэкспортируемый тип ключа: значение такого типа может
// создать только этот пакет, поэтому чужой код ключ не подделает.
type requestIDKey struct{}

// WithRequestID кладёт идентификатор запроса в контекст.
// Прочитать его можно только через RequestID — никакой другой код,
// даже знающий строку "request_id", до него не доберётся.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID достаёт идентификатор запроса. ok == false, если его нет.
func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok
}
