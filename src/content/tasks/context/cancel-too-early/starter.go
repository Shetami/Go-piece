package main

import (
	"context"
	"fmt"
	"time"
)

// requestCtx создаёт контекст запроса с таймаутом в секунду.
func requestCtx(parent context.Context) context.Context {
	ctx, cancel := context.WithTimeout(parent, time.Second)
	defer cancel()
	return ctx
}

func query(ctx context.Context) error {
	select {
	case <-time.After(10 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	ctx := requestCtx(context.Background())
	fmt.Println("запрос:", query(ctx))
}
