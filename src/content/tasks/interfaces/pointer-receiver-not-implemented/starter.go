package main

import "fmt"

type Shape interface{ Area() float64 }

type Circle struct{ R float64 }

func (c *Circle) Area() float64 { return 3 * c.R * c.R }

type Square struct{ A float64 }

func (s Square) Area() float64 { return s.A * s.A }

func main() {
	shapes := []Shape{Circle{R: 1}, Square{A: 2}}

	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	fmt.Println("площадь:", total)
}
