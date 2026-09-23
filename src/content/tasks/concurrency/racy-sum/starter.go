package main

import (
	"fmt"
	"sync"
)

func main() {
	total := 0
	var wg sync.WaitGroup

	for i := 1; i <= 1000; i++ {
		wg.Go(func() {
			total += i
		})
	}
	wg.Wait()

	fmt.Println("сумма:", total)
}
