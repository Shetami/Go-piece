package main

import (
	"fmt"
	"strings"
)

// slug превращает заголовок в кусок URL: пробелы — в дефисы, всё в нижний регистр.
func slug(title string) string {
	// ReplaceAll — то же самое, что Replace с n = -1: заменить все вхождения.
	s := strings.ReplaceAll(title, " ", "-")
	return strings.ToLower(s)
}

func main() {
	fmt.Println(slug("Go"))
	fmt.Println(slug("Hello World"))
	fmt.Println(slug("Go Is Fun Again"))
}
