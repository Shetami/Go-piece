package main

import (
	"fmt"
	"sync"
)

func main() {
	jobs := make(chan int)
	defer close(jobs)

	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		total int
	)
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				mu.Lock()
				total += j
				mu.Unlock()
			}
		}()
	}

	for i := 1; i <= 10; i++ {
		jobs <- i
	}
	wg.Wait()
	fmt.Println("сумма:", total)
}
