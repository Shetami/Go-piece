package main

import (
	"context"
	"fmt"
	"time"
)

func query(ctx context.Context) error {
	select {
	case <-time.After(2 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func handler(ctx context.Context) error {
	return query(context.Background())
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := handler(ctx)
	fmt.Println(err, time.Since(start).Round(time.Second))
}
