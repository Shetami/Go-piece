package main

import (
	"fmt"
	"strconv"
)

// Parse разбирает строку как число нужного типа.
func Parse[T int | float64](s string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case int:
		n, err := strconv.Atoi(s)
		return T(n), err
	default:
		f, err := strconv.ParseFloat(s, 64)
		return T(f), err
	}
}

func main() {
	// T встречается только в результате — вывести его не из чего, указываем явно.
	port, err := Parse[int]("8080")
	if err != nil {
		panic(err)
	}
	ratio, err := Parse[float64]("0.75")
	if err != nil {
		panic(err)
	}
	fmt.Println(port+1, ratio*2)
}
