package main

import (
	"fmt"
	"sync"
)

type Stats struct {
	mu   sync.RWMutex
	hits map[string]int
}

func (s *Stats) Inc(page string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.hits[page]++
}

func (s *Stats) Get(page string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hits[page]
}

func main() {
	s := &Stats{hits: map[string]int{}}

	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() { s.Inc("/") })
	}
	wg.Wait()

	fmt.Println("просмотров:", s.Get("/"))
}
