package main

import (
	"context"
	"fmt"
	"time"
)

type key struct{}

func saveAudit(ctx context.Context, msg string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
		fmt.Printf("аудит [%s]: %s\n", ctx.Value(key{}), msg)
		return nil
	}
}

func handle(ctx context.Context) {
	<-ctx.Done() // клиент ушёл, запрос прерван

	// Аудит нужен именно в этом случае — и он должен записаться.
	// WithoutCancel сохраняет значения запроса, но отрезает его отмену.
	if err := saveAudit(context.WithoutCancel(ctx), "запрос прерван"); err != nil {
		fmt.Println("аудит не записан:", err)
	}
}

func main() {
	ctx := context.WithValue(context.Background(), key{}, "r42")
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	handle(ctx)
}
