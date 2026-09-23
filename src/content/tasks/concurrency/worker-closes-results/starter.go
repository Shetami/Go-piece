package main

import "fmt"

func main() {
	jobs := make(chan int)
	results := make(chan int)

	for range 3 {
		go func() {
			for j := range jobs {
				results <- j * j
			}
			close(results)
		}()
	}

	go func() {
		for i := 1; i <= 5; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println("сумма квадратов:", sum)
}
