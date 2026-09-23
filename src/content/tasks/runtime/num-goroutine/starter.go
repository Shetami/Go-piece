package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Println(runtime.NumGoroutine())

	block := make(chan struct{})
	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() {
			wg.Done()
			<-block
		}()
	}
	wg.Wait()
	fmt.Println(runtime.NumGoroutine())

	close(block)
	time.Sleep(10 * time.Millisecond)
	fmt.Println(runtime.NumGoroutine())
}
