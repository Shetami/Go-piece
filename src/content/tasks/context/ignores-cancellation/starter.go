package main

import (
	"context"
	"fmt"
	"time"
)

// fetch делает пять шагов по 400 мс.
// Контракт: если контекст отменили — вернуть ctx.Err() и не доделывать шаги.
func fetch(ctx context.Context) (string, error) {
	for step := 1; step <= 5; step++ {
		time.Sleep(400 * time.Millisecond)
		fmt.Println("шаг", step)
	}
	return "готово", nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := fetch(ctx)
	if err != nil {
		fmt.Println("прервано:", err)
		return
	}
	fmt.Println("результат:", res)
}
