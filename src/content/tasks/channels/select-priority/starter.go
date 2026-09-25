package main

import "fmt"

func main() {
	high := make(chan string, 10)
	low := make(chan string, 10)
	for i := 1; i <= 10; i++ {
		high <- fmt.Sprint("срочно-", i)
		low <- fmt.Sprint("обычно-", i)
	}
	close(high)
	close(low)

	// Срочные сообщения должны обрабатываться раньше обычных.
	for high != nil || low != nil {
		select {
		case m, ok := <-high:
			if !ok {
				high = nil
				continue
			}
			fmt.Println(m)
		case m, ok := <-low:
			if !ok {
				low = nil
				continue
			}
			fmt.Println(m)
		}
	}
}
