package main

import (
	"fmt"
	"strings"
)

// slug превращает заголовок в кусок URL: пробелы — в дефисы, всё в нижний регистр.
func slug(title string) string {
	s := strings.Replace(title, " ", "-", 1)
	return strings.ToLower(s)
}

func main() {
	fmt.Println(slug("Go"))
	fmt.Println(slug("Hello World"))
	fmt.Println(slug("Go Is Fun Again"))
}
