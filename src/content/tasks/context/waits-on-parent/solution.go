package main

import (
	"context"
	"fmt"
	"time"
)

// slowSearch не умеет останавливаться и всегда работает секунду.
func slowSearch(q string) string {
	time.Sleep(time.Second)
	return "результаты для " + q
}

// search ждёт ответа не дольше 100 мс.
func search(ctx context.Context, q string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	res := make(chan string, 1)
	go func() { res <- slowSearch(q) }()

	// Ждём того контекста, на который поставили таймаут.
	select {
	case r := <-res:
		return r, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	start := time.Now()
	r, err := search(context.Background(), "go")
	fmt.Printf("%q %v за %v\n", r, err, time.Since(start).Round(100*time.Millisecond))
}
