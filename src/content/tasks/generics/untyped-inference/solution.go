package main

import "fmt"

type Celsius float64

func Pick[T any](a, b T) T { return b }

func main() {
	v := Pick(1, 2.5)
	fmt.Printf("%T %v\n", v, v)

	r := Pick(1, 'a')
	fmt.Printf("%T %v\n", r, r)

	var n int8 = 5
	x := Pick(n, 100)
	fmt.Printf("%T %v\n", x, x)

	c := Pick(Celsius(36.6), 2)
	fmt.Printf("%T %v\n", c, c)
}
