package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("apple" < "banana", "Zebra" < "apple", "10" < "9")
	fmt.Println(strings.Index("привет", "в"), strings.IndexRune("привет", 'в'))
	fmt.Println(strings.ToUpper("straße"), len("straße"))
	fmt.Println(strings.EqualFold("Go", "GO"), "Go" == "GO")
}
