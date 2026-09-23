package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(len(strings.Split("a,b,,c", ",")))
	fmt.Println(len(strings.Split("", ",")))
	fmt.Println(len(strings.Split("abc", "")))
	fmt.Println(len(strings.Fields("  a  b  ")))
	fmt.Printf("%q\n", strings.Split("a,b,", ","))

	fmt.Printf("%q\n", strings.TrimLeft("xxaxbx", "x"))
	fmt.Printf("%q\n", strings.TrimPrefix("xxaxbx", "x"))
	fmt.Printf("%q\n", strings.Trim("[[id]]", "[]"))
}
