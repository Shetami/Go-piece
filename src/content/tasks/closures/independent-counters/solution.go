package main

import "fmt"

func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func main() {
	a := counter()
	b := counter()
	fmt.Println(a(), a(), b(), a())

	c := a
	fmt.Println(c(), a())
}
