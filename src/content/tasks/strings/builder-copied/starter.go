package main

import (
	"fmt"
	"strings"
)

func writeHeader(b strings.Builder) {
	b.WriteString("== отчёт ==\n")
}

func main() {
	var b strings.Builder
	b.WriteString("начало\n")
	writeHeader(b)
	b.WriteString("конец\n")
	fmt.Print(b.String())
}
