package main

import "fmt"

type Point struct{ X, Y int }

func (p *Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func main() {
	p := Point{1, 2}
	fmt.Println(p)
	fmt.Println([]Point{{3, 4}})
	fmt.Printf("точка %v\n", p)
}
