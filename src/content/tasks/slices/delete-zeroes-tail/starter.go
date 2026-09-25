package main

import (
	"fmt"
	"slices"
)

func main() {
	s := []int{1, 2, 3, 4}
	t := slices.Delete(s, 1, 2)

	fmt.Println(t, len(t), cap(t))
	fmt.Println(s)
}
