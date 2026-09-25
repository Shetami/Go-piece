package main

import "fmt"

func main() {
	n := 0
	next := func() int {
		n++
		return n
	}

	fmt.Println(next(), next(), next())

	a := []int{0, 0, 0}
	i := 0
	a[i], i = next(), 2
	fmt.Println(a, i)

	x, y := 1, 2
	x, y = y, x+y
	fmt.Println(x, y)
}
