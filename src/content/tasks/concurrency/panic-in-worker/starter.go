package main

import (
	"fmt"
	"sync"
)

func work(x int) int {
	if x == 3 {
		panic("плохие данные")
	}
	return x * x
}

func main() {
	results := make([]string, 5)

	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = fmt.Sprint(work(i))
		}()
	}
	wg.Wait()

	fmt.Println(results)
}
