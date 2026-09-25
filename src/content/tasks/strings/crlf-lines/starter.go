package main

import (
	"fmt"
	"strings"
)

func main() {
	// Файл пришёл из Windows: строки кончаются на \r\n.
	data := "admin\r\nguest\r\nroot\r\n"
	allowed := map[string]bool{"admin": true, "root": true}

	for _, name := range strings.Split(strings.TrimSpace(data), "\n") {
		fmt.Printf("%q: %v\n", name, allowed[name])
	}
}
