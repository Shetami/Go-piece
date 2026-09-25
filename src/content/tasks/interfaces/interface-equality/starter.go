package main

import "fmt"

type ID int

type point struct{ x, y int }

func main() {
	var a any = 1
	var b any = 1
	var c any = int64(1)
	var d any = ID(1)

	fmt.Println(a == b)
	fmt.Println(a == c)
	fmt.Println(a == 1)
	fmt.Println(a == d)

	var s, t any = point{1, 2}, point{1, 2}
	fmt.Println(s == t)

	var p, q any = &point{1, 2}, &point{1, 2}
	fmt.Println(p == q)
}
