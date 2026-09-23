package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
)

func main() {
	texts := []string{"go is fun", "go is fast", "fun is fun"}
	counts := map[string]int{}
	var mu sync.Mutex

	var wg sync.WaitGroup
	for _, t := range texts {
		wg.Go(func() {
			// Каждая горутина считает у себя, а в общую мапу сливает один раз
			// под мьютексом: так блокировка берётся трижды, а не 9000 раз.
			local := map[string]int{}
			for range 1000 {
				for _, w := range strings.Fields(t) {
					local[w]++
				}
			}
			mu.Lock()
			for w, n := range local {
				counts[w] += n
			}
			mu.Unlock()
		})
	}
	wg.Wait()

	for _, w := range slices.Sorted(maps.Keys(counts)) {
		fmt.Println(w, counts[w])
	}
}
