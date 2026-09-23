package main

import "fmt"

func main() {
	s := "Go — это круто"

	fmt.Println(len(s))
	fmt.Println(len([]rune(s)))

	for i, r := range "Gö" {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()

	fmt.Println(s[3])
	fmt.Printf("%c\n", s[0])
}
