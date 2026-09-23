package main

import (
	"fmt"
	"unicode"
)

// countUpper считает заглавные буквы.
func countUpper(s string) int {
	n := 0
	// range по строке отдаёт руны — целые символы, а не отдельные байты.
	for _, r := range s {
		if unicode.IsUpper(r) {
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
