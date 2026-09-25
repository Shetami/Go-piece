package main

import (
	"fmt"
	"strings"
)

// Replacer проходит строку один раз и не заменяет то, что сам же вставил.
var htmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

// escape готовит текст для вставки в HTML.
func escape(s string) string {
	return htmlEscaper.Replace(s)
}

func main() {
	fmt.Println(escape("<b>Tom & Jerry</b>"))
}
