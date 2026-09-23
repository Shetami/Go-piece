package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var wg sync.WaitGroup
	var done atomic.Int64

	for range 5 {
		go func() {
			wg.Add(1)
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			done.Add(1)
		}()
	}

	wg.Wait()
	fmt.Println("готово задач:", done.Load())
}
