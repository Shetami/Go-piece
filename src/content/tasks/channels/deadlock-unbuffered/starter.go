package main

import "fmt"

func main() {
	results := make(chan int)

	for i := 1; i <= 3; i++ {
		results <- i * i
	}
	close(results)

	sum := 0
	for v := range results {
		sum += v
	}

	fmt.Println("сумма квадратов:", sum)
}
