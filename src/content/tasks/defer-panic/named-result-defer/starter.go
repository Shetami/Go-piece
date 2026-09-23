package main

import "fmt"

func counter() (count int) {
	defer func() {
		count *= 2
	}()

	for i := 0; i < 3; i++ {
		defer fmt.Println("defer", i)
		count++
	}

	return count + 1
}

func main() {
	fmt.Println("итог:", counter())
}
