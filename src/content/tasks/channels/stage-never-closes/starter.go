package main

import "fmt"

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

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func main() {
	sum := 0
	for v := range square(gen(1, 2, 3)) {
		sum += v
	}
	fmt.Println("сумма квадратов:", sum)
}
