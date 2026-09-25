package main

import "fmt"

func main() {
	base := []int{1, 2, 3, 4, 5}

	a := base[1:3]
	a = append(a, 100)

	b := base[1:3:3]
	b = append(b, 200)

	fmt.Println("base:", base)
	fmt.Println("a:", a)
	fmt.Println("b:", b)
}
