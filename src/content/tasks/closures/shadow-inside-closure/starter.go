package main

import "fmt"

func main() {
	x := 1

	set := func() {
		x := 2
		_ = x
	}
	set()
	fmt.Println(x)

	inc := func() { x++ }
	inc()
	inc()
	fmt.Println(x)

	reset := func(x int) { x = 0 }
	reset(x)
	fmt.Println(x)

	next := func() func() int {
		x := x * 10
		return func() int {
			x++
			return x
		}
	}()
	fmt.Println(next(), next(), x)
}
