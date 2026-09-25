package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func main() {
	s := "привет, мир"

	fmt.Println(len(s))

	i := strings.Index(s, "мир")
	fmt.Println(i)
	fmt.Println(s[i:])
	fmt.Println(utf8.RuneCountInString(s[:i]))
}
