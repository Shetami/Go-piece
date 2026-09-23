package main

import (
	"fmt"
	"sync"
)

func main() {
	total := 0
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	for i := 1; i <= 1000; i++ {
		wg.Go(func() {
			// Чтение и запись total — одна операция под мьютексом.
			mu.Lock()
			total += i
			mu.Unlock()
		})
	}
	wg.Wait()

	fmt.Println("сумма:", total)
}
