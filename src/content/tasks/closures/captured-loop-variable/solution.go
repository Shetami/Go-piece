package main

import "fmt"

func makeHandlers() []func() {
	var handlers []func()

	for i := 0; i < 3; i++ {
		handlers = append(handlers, func() {
			fmt.Print(i, " ")
		})
	}

	return handlers
}

func main() {
	for _, h := range makeHandlers() {
		h()
	}
	fmt.Println()

	x := 10
	inc := func() { x++ }
	inc()
	inc()
	fmt.Println("x =", x)
}
