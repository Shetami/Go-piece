package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var a []int
	b := []int{}

	fmt.Println(a == nil, b == nil)
	fmt.Println(len(a), len(b))
	fmt.Println(a, b)

	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	fmt.Println(string(ja), string(jb))

	a = append(a, 1)
	fmt.Println(a)
}
