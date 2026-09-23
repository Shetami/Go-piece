package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func worker(id int, sum *int64, wg sync.WaitGroup) {
	defer wg.Done()
	atomic.AddInt64(sum, int64(id))
}

func main() {
	var wg sync.WaitGroup
	var sum int64

	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go worker(i, &sum, wg)
	}

	wg.Wait()
	fmt.Println("сумма:", sum)
}
