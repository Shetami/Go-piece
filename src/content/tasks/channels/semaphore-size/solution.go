package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const limit = 2 // не больше двух одновременных запросов к API

func main() {
	urls := make([]string, 10)
	// Ёмкость буфера и есть число одновременно работающих.
	sem := make(chan struct{}, limit)

	var active, peak atomic.Int32
	var wg sync.WaitGroup
	for range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			n := active.Add(1)
			for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
			}
			time.Sleep(10 * time.Millisecond)
			active.Add(-1)
		}()
	}
	wg.Wait()
	fmt.Println("одновременно:", peak.Load(), "лимит:", limit)
}
