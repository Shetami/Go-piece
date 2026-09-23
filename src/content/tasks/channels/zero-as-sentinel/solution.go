package main

import "fmt"

func main() {
	readings := make(chan int, 10)
	for _, t := range []int{5, 3, 0, 7} {
		readings <- t
	}
	close(readings)

	sum := 0
	// range сам заканчивается на закрытии канала, а не на нулевом значении.
	for v := range readings {
		sum += v
	}
	fmt.Println("сумма показаний:", sum)
}
