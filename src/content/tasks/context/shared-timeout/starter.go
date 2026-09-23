package main

import (
	"context"
	"fmt"
	"time"
)

func fetch(ctx context.Context, url string) error {
	select {
	case <-time.After(100 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// На каждый запрос — не больше 150 мс.
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	for _, url := range []string{"a.example", "b.example", "c.example"} {
		fmt.Println(url, fetch(ctx, url))
	}
}
