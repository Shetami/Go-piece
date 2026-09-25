package main

import (
	"fmt"
	"strings"
)

func main() {
	s := " a  b "

	parts := strings.Split(s, " ")
	fmt.Printf("%d %q\n", len(parts), parts)

	fields := strings.Fields(s)
	fmt.Printf("%d %q\n", len(fields), fields)

	empty := strings.Split("", ",")
	fmt.Printf("%d %q\n", len(empty), empty)
}
