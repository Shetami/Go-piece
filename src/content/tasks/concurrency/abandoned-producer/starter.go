package main

import (
	"fmt"
	"runtime"
	"time"
)

// numbers отдаёт 0, 1, 2, ... бесконечно.
func numbers() <-chan int {
	ch := make(chan int)
	go func() {
		for i := 0; ; i++ {
			ch <- i
		}
	}()
	return ch
}

func firstN(n int) []int {
	var out []int
	for v := range numbers() {
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
