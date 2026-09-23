package main

import (
	"fmt"
	"sync"
)

func process(id int) string {
	if id == 1 {
		panic("сбой в задаче 1")
	}
	return fmt.Sprint("задача ", id, ": ок")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("поймали:", r)
		}
	}()

	results := make([]string, 3)
	var wg sync.WaitGroup
	for i := range 3 {
		wg.Go(func() {
			results[i] = process(i)
		})
	}
	wg.Wait()

	for _, r := range results {
		fmt.Println(r)
	}
}
