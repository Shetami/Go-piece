package main

import "fmt"

type Number interface {
	~int | ~float64 // тильда: любой тип, построенный на int или float64
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
	fmt.Println(Sum(Celsius(20), Celsius(1.5)))
}
