package main

import (
	"fmt"
	"strings"
)

// query собирает строку запроса. Одинаковые параметры должны давать
// одинаковую строку: по ней ищут в кэше.
func query(params map[string]string) string {
	var b strings.Builder
	for k, v := range params {
		fmt.Fprintf(&b, "%s=%s&", k, v)
	}
	return strings.TrimSuffix(b.String(), "&")
}

func main() {
	params := map[string]string{"q": "go", "page": "2", "lang": "ru", "sort": "new"}
	for range 3 {
		fmt.Println(query(params))
	}
}
