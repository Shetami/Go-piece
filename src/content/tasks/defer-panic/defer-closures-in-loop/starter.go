package main

import "fmt"

func main() {
	for i := range 3 {
		defer func() { fmt.Println("i =", i) }()
	}

	x := 0
	for range 3 {
		x++
		defer func() { fmt.Println("x =", x) }()
	}

	fmt.Println("конец main")
}
