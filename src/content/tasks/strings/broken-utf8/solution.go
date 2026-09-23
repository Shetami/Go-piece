package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "я"
	broken := s[:1]

	for i, r := range broken {
		fmt.Println(i, r, r == utf8.RuneError)
	}
	fmt.Println(len(broken), utf8.ValidString(broken), utf8.RuneCountInString(broken))
	fmt.Println([]byte(s))
}
