package main

import (
	"context"
	"fmt"
	"time"
)

// requestCtx создаёт контекст запроса с таймаутом в секунду.
// cancel отдаём вызывающему: отменять контекст должен тот, кто им пользуется.
func requestCtx(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, time.Second)
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
	ctx, cancel := requestCtx(context.Background())
	defer cancel()
	fmt.Println("запрос:", query(ctx))
}
