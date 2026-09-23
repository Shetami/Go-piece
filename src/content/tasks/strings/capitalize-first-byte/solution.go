package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// capitalize делает первую букву заглавной.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	// Первая буква — это первая руна, и её ширина в байтах бывает разной.
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func main() {
	fmt.Printf("%q\n", capitalize("go"))
	fmt.Printf("%q\n", capitalize("язык"))
}
