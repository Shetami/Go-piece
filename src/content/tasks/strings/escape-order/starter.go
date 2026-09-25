package main

import (
	"fmt"
	"strings"
)

// escape готовит текст для вставки в HTML.
func escape(s string) string {
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "&", "&amp;")
	return s
}

func main() {
	fmt.Println(escape("<b>Tom & Jerry</b>"))
}
