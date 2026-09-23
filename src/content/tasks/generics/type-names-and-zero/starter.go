package main

import "fmt"

type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func Zero[T any]() T {
	var z T
	return z
}

func main() {
	fmt.Printf("%T\n", Pair[string, int]{})
	fmt.Printf("%+v\n", Pair[string, []int]{Key: "a"})
	fmt.Printf("%q %v %v\n", Zero[string](), Zero[*int]() == nil, Zero[[]int]() == nil)
	fmt.Println(Zero[Pair[int, bool]]())
}
