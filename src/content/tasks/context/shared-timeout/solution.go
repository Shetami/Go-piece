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

// fetchWithTimeout даёт каждому запросу свои 150 мс. Отдельная функция —
// чтобы defer cancel срабатывал после каждого запроса, а не в конце main.
func fetchWithTimeout(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	return fetch(ctx, url)
}

// На каждый запрос — не больше 150 мс.
func main() {
	for _, url := range []string{"a.example", "b.example", "c.example"} {
		fmt.Println(url, fetchWithTimeout(url))
	}
}
