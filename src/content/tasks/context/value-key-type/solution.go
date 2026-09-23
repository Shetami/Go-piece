package main

import (
	"context"
	"fmt"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func logf(ctx context.Context, msg string) {
	// Ключ того же типа, что при записи: строка "request_id" и ctxKey — разные ключи.
	id, _ := ctx.Value(requestIDKey).(string)
	if id == "" {
		id = "без id"
	}
	fmt.Printf("[%s] %s\n", id, msg)
}

func main() {
	ctx := withRequestID(context.Background(), "a1b2")
	logf(ctx, "заказ создан")
}
