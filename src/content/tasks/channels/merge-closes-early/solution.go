package main

import (
	"fmt"
	"sync"
)

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// merge сливает несколько каналов в один и закрывает его,
// когда закончатся все входы.
func merge(ins ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(ins))
	for _, in := range ins {
		go func() {
			defer wg.Done()
			for v := range in {
				out <- v
			}
		}()
	}
	// Закрывает один — тот, кто знает, что все писатели закончили.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	sum := 0
	for v := range merge(gen(1, 2, 3), gen(10, 20), gen(100)) {
		sum += v
	}
	fmt.Println("сумма:", sum)
}
