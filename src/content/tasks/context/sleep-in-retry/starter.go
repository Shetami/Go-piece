package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func call() error { return errors.New("сервис недоступен") }

func callWithRetry(ctx context.Context) error {
	for attempt := 1; attempt <= 5; attempt++ {
		fmt.Println("попытка", attempt)
		if err := call(); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("попытки кончились")
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	fmt.Println("итог:", callWithRetry(ctx))
}
