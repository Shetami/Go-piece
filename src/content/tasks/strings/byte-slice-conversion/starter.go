package main

import (
	"fmt"
	"strconv"
)

func main() {
	s := "hello"
	b := []byte(s)
	b[0] = 'J'
	fmt.Println(s, string(b))

	r := []rune("привет")
	r[0] = 'П'
	fmt.Println(string(r), len(r), len(string(r)))

	fmt.Println(string(rune(1071)), strconv.Itoa(1071))
}
