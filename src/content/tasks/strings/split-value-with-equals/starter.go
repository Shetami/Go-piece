package main

import (
	"fmt"
	"strings"
)

// parseLine разбирает строку конфига вида «ключ = значение».
func parseLine(line string) (key, value string) {
	parts := strings.Split(line, "=")
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
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
