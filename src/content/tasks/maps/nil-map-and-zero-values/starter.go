package main

import "fmt"

type Counter struct {
	Hits int
}

func main() {
	var m map[string]int
	fmt.Println(m == nil, len(m), m["нет такого"])

	v, ok := m["нет такого"]
	fmt.Println(v, ok)

	counts := map[string]int{}
	counts["a"]++
	counts["a"]++
	fmt.Println(counts["a"], counts["b"])

	stats := map[string]*Counter{"x": {}}
	stats["x"].Hits++
	fmt.Println(stats["x"].Hits)
}
