package main

import "fmt"

type point struct{ x, y int }

func main() {
	pc := make(chan point, 1)
	p := point{1, 2}
	pc <- p
	p.x = 100
	fmt.Println(<-pc)

	sc := make(chan []int, 1)
	s := []int{1, 2, 3}
	sc <- s
	s[0] = 100
	fmt.Println(<-sc)

	mc := make(chan map[string]int, 1)
	m := map[string]int{"a": 1}
	mc <- m
	m["a"] = 100
	fmt.Println(<-mc)
}
