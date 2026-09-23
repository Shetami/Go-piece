package main

import (
	"fmt"
	"strings"
)

// writeHeader получает указатель: писать надо в тот же Builder,
// а копировать непустой Builder запрещено.
func writeHeader(b *strings.Builder) {
	b.WriteString("== отчёт ==\n")
}

func main() {
	var b strings.Builder
	b.WriteString("начало\n")
	writeHeader(&b)
	b.WriteString("конец\n")
	fmt.Print(b.String())
}
