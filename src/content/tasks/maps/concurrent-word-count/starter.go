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

	var wg sync.WaitGroup
	for _, t := range texts {
		wg.Go(func() {
			for range 1000 {
				for _, w := range strings.Fields(t) {
					counts[w]++
				}
			}
		})
	}
	wg.Wait()

	for _, w := range slices.Sorted(maps.Keys(counts)) {
		fmt.Println(w, counts[w])
	}
}
