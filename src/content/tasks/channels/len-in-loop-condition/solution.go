package main

import "fmt"

func main() {
	ch := make(chan int, 10)
	for i := 1; i <= 6; i++ {
		ch <- i
	}
	close(ch)

	sum := 0
	// range читает до закрытия канала — не нужно знать, сколько там значений.
	for v := range ch {
		sum += v
	}
	fmt.Println("сумма:", sum)
}
