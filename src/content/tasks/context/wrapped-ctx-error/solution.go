package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func call(ctx context.Context) error {
	select {
	case <-time.After(50 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return fmt.Errorf("запрос к базе: %w", ctx.Err())
	}
}

func status(err error) string {
	// errors.Is заглядывает под обёртки, == сравнивает только верхнее звено.
	switch {
	case err == nil:
		return "200"
	case errors.Is(err, context.DeadlineExceeded):
		return "504 таймаут"
	case errors.Is(err, context.Canceled):
		return "499 клиент ушёл"
	default:
		return "500"
	}
}

func main() {
	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel1()
	fmt.Println(status(call(ctx1)))

	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()
	fmt.Println(status(call(ctx2)))

	fmt.Println(status(call(context.Background())))
}
