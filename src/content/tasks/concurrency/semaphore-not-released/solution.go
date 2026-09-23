package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	sem := make(chan struct{}, 2) // не больше двух задач одновременно
	var wg sync.WaitGroup
	var ok, failed atomic.Int64

	for job := 1; job <= 20; job++ {
		wg.Go(func() {
			sem <- struct{}{}
			// Слот освобождается на любом выходе из функции, включая ранний return.
			defer func() { <-sem }()
			if job%3 == 0 {
				failed.Add(1)
				return
			}
			ok.Add(1)
		})
	}
	wg.Wait()

	fmt.Println("успешно:", ok.Load(), "с ошибкой:", failed.Load())
}
