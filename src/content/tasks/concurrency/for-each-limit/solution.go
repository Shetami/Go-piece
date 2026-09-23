package main

import "sync"

// ForEachLimit вызывает f для каждого элемента, одновременно — не больше limit
// вызовов. Возвращается, когда все вызовы закончились.
func ForEachLimit(items []int, limit int, f func(int)) {
	// Семафор на буферизованном канале: занять место — отправить, освободить — прочитать.
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for _, it := range items {
		// Занимаем место до запуска горутины: так их никогда не бывает
		// больше limit, даже ждущих.
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			f(it)
		})
	}
	wg.Wait()
}
