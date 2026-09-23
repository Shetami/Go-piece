package main

import (
	"fmt"
	"unicode"
)

// countUpper считает заглавные буквы.
func countUpper(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if unicode.IsUpper(rune(s[i])) {
			n++
		}
	}
	return n
}

func main() {
	fmt.Println(countUpper("Go"))
	fmt.Println(countUpper("Привет, Мир"))
	fmt.Println(countUpper("Élan"))
}
