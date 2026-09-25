package main

import (
	"fmt"
	"strconv"
)

func main() {
	n := 1071
	fmt.Println(string(rune(n)))
	fmt.Println(strconv.Itoa(n))

	fmt.Println(string(rune(65)) + strconv.Itoa(65))

	b := []byte("Go")
	b = append(b, 33)
	fmt.Println(string(b))

	fmt.Println('A')
	fmt.Println(string(rune(-1)))
}
