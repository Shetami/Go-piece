package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var n atomic.Int64
	var wg sync.WaitGroup

	for range 50 {
		wg.Go(func() {
			for range 100 {
				n.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println(n.Load())

	old := n.Swap(0)
	fmt.Println(old, n.Load())

	fmt.Println(n.CompareAndSwap(1, 5), n.Load())
	fmt.Println(n.CompareAndSwap(0, 5), n.Load())
}
