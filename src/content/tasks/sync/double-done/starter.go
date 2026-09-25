package main

import (
	"fmt"
	"slices"
	"sync"
)

func fetch(id int, wg *sync.WaitGroup, results chan<- string) {
	defer wg.Done()
	if id%2 == 0 {
		results <- fmt.Sprintf("%d: ошибка", id)
		wg.Done()
		return
	}
	results <- fmt.Sprintf("%d: ok", id)
}

func main() {
	results := make(chan string, 4)

	var wg sync.WaitGroup
	for id := 1; id <= 4; id++ {
		wg.Add(1)
		go fetch(id, &wg, results)
	}
	wg.Wait()
	close(results)

	var lines []string
	for r := range results {
		lines = append(lines, r)
	}
	slices.Sort(lines)
	for _, l := range lines {
		fmt.Println(l)
	}
}
