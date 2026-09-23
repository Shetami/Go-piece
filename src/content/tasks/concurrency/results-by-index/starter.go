package main

import (
	"fmt"
	"sync"
)

func main() {
	words := []string{"го", "rust", "c"}
	lens := make([]int, len(words))

	var wg sync.WaitGroup
	for i, w := range words {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lens[i] = len(w)
		}()
	}
	wg.Wait()

	fmt.Println(lens)
}
