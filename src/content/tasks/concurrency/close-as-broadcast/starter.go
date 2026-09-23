package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	start := make(chan struct{})
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		order []int
	)

	for i := range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			mu.Lock()
			order = append(order, i)
			mu.Unlock()
		}()
	}

	close(start)
	wg.Wait()

	sort.Ints(order)
	fmt.Println(order, len(order))

	v, ok := <-start
	fmt.Println(v, ok)
}
