package main

import "fmt"

func main() {
	readings := make(chan int, 10)
	for _, t := range []int{5, 3, 0, 7} {
		readings <- t
	}
	close(readings)

	sum := 0
	for {
		v := <-readings
		if v == 0 { // канал закрыт
			break
		}
		sum += v
	}
	fmt.Println("сумма показаний:", sum)
}
