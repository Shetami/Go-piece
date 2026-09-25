package main

import (
	"fmt"
	"strings"
)

// parseLine разбирает строку конфига вида «ключ = значение».
func parseLine(line string) (key, value string) {
	// Режем только по первому «=»: в значении он тоже может встретиться.
	k, v, _ := strings.Cut(line, "=")
	return strings.TrimSpace(k), strings.TrimSpace(v)
}

func main() {
	config := []string{
		"user = admin",
		"dsn = postgres://app@db/shop?sslmode=disable",
		"motd = 2+2=4",
	}
	for _, line := range config {
		k, v := parseLine(line)
		fmt.Printf("%s → %q\n", k, v)
	}
}
