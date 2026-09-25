package main

import "fmt"

func main() {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	sum, iters := 0, 0
	for iters < 10 {
		iters++
		select {
		case v, ok := <-ch:
			if !ok {
				break
			}
			sum += v
		}
	}
	fmt.Println("без метки:", sum, iters)

	ch = make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	sum, iters = 0, 0
loop:
	for iters < 10 {
		iters++
		select {
		case v, ok := <-ch:
			if !ok {
				break loop
			}
			sum += v
		}
	}
	fmt.Println("с меткой:", sum, iters)
}
