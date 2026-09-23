package main

import (
	"fmt"
	"sync"
)

func main() {
	jobs := make(chan int)
	results := make(chan int)

	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			for j := range jobs {
				results <- j * j
			}
		})
	}

	// results закрывает не воркер, а тот, кто знает, что воркеров больше нет.
	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		for i := 1; i <= 5; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println("сумма квадратов:", sum)
}
