package main

import (
	"fmt"
	"strings"
)

// capitalize делает первую букву заглавной.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func main() {
	fmt.Printf("%q\n", capitalize("go"))
	fmt.Printf("%q\n", capitalize("язык"))
}
