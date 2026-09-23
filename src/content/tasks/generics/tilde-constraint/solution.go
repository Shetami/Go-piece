package main

import "fmt"

type Number interface {
	~int | ~float64
}

func Sum[T Number](xs ...T) T {
	var s T
	for _, x := range xs {
		s += x
	}
	return s
}

type Celsius float64

func main() {
	fmt.Println(Sum(1, 2, 3))
	fmt.Println(Sum(1.5, 2))

	r := Sum(Celsius(10), 20)
	fmt.Printf("%v %T\n", r, r)

	fmt.Printf("%v %T\n", Sum[float64](), Sum[float64]())
}
