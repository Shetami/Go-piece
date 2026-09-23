package main

import "fmt"

func main() {
	x := 1
	defer fmt.Println("аргумент:", x)
	defer func() {
		fmt.Println("замыкание:", x)
	}()

	x = 2
	fmt.Println("тело:", x)
}
