package main

import "fmt"

func main() {
	ch := make(chan int, 10)
	for i := 1; i <= 6; i++ {
		ch <- i
	}
	close(ch)

	sum := 0
	for i := 0; i < len(ch); i++ {
		sum += <-ch
	}
	fmt.Println("сумма:", sum)
}
