package main

import "fmt"

func fill(m map[string]int) {
	m["a"] = 1
}

func replace(m map[string]int) {
	m = map[string]int{"b": 2}
	m["c"] = 3
}

func main() {
	m := map[string]int{}
	fill(m)
	replace(m)
	fmt.Println(m, len(m))

	m["hits"]++
	fmt.Println(m)
}
