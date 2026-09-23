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
	id, _ := ctx.Value("request_id").(string)
	if id == "" {
		id = "без id"
	}
	fmt.Printf("[%s] %s\n", id, msg)
}

func main() {
	ctx := withRequestID(context.Background(), "a1b2")
	logf(ctx, "заказ создан")
}
