package main

import (
	"fmt"
	"math"
)

func main() {
	m := map[float64]string{}
	nan := math.NaN()

	m[nan] = "первый"
	m[nan] = "второй"
	m[1] = "один"

	fmt.Println(len(m))

	_, ok := m[nan]
	fmt.Println(ok)

	delete(m, nan)
	fmt.Println(len(m))

	clear(m)
	fmt.Println(len(m))
}
