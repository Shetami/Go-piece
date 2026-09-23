package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var hits sync.Map
	var wg sync.WaitGroup

	for range 1000 {
		wg.Go(func() {
			// В мапе лежит атомарный счётчик: LoadOrStore гарантирует, что он один
			// на ключ, а Add увеличивает его без гонки «прочитал — записал».
			v, _ := hits.LoadOrStore("/", new(atomic.Int64))
			v.(*atomic.Int64).Add(1)
		})
	}
	wg.Wait()

	v, _ := hits.Load("/")
	fmt.Println("просмотров:", v.(*atomic.Int64).Load())
}
