package main

import "fmt"

func main() {
	a := make(chan int)
	b := make(chan int)

	go func() {
		for _, v := range []int{1, 2} {
			a <- v
		}
		close(a)
	}()
	go func() {
		b <- 10
		close(b)
	}()

	sum, closed := 0, 0
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil
				closed++
				continue
			}
			sum += v
		case v, ok := <-b:
			if !ok {
				b = nil
				closed++
				continue
			}
			sum += v
		}
	}

	fmt.Println(sum, closed)
}
