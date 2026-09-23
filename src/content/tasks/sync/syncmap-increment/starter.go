package main

import (
	"fmt"
	"sync"
)

func main() {
	var hits sync.Map
	var wg sync.WaitGroup

	for range 1000 {
		wg.Go(func() {
			v, _ := hits.LoadOrStore("/", 0)
			hits.Store("/", v.(int)+1)
		})
	}
	wg.Wait()

	v, _ := hits.Load("/")
	fmt.Println("просмотров:", v)
}
