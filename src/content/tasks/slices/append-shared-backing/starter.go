package main

import "fmt"

func main() {
	s := make([]int, 3, 5)
	a := append(s, 1)
	b := append(s, 2)

	fmt.Println(a, b)
	fmt.Println(len(s), cap(s))
	fmt.Println(len(a), cap(a))
}
