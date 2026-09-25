package main

import "fmt"

func main() {
	m := map[string][]int{}

	v := m["x"]
	v = append(v, 1)
	fmt.Println(len(m), m["x"] == nil, len(v))

	m["y"] = append(m["y"], 1)
	m["y"] = append(m["y"], 2)
	fmt.Println(len(m), m["y"])

	counts := map[string]int{}
	counts["go"]++
	counts["go"]++
	fmt.Println(counts)
}
