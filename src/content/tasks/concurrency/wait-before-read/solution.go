package main

import (
	"fmt"
	"sync"
)

func main() {
	results := make(chan int)

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Go(func() { results <- i * 10 })
	}

	// Ждём и закрываем в отдельной горутине, чтобы main сразу начал читать:
	// иначе отправители ждут читателя, а читатель ждёт отправителей.
	go func() {
		wg.Wait()
		close(results)
	}()

	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println("сумма:", sum)
}
