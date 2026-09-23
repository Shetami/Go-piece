package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// numbers отдаёт 0, 1, 2, ... пока ctx не отменят.
func numbers(ctx context.Context) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 0; ; i++ {
			select {
			case ch <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

func firstN(n int) []int {
	// Отмена при выходе из функции останавливает производителя,
	// даже когда мы бросили читать на середине.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var out []int
	for v := range numbers(ctx) {
		if len(out) == n {
			break
		}
		out = append(out, v)
	}
	return out
}

func main() {
	for range 3 {
		fmt.Println(firstN(3))
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println("горутин:", runtime.NumGoroutine())
}
