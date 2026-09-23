package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// query собирает строку запроса. Одинаковые параметры должны давать
// одинаковую строку: по ней ищут в кэше.
func query(params map[string]string) string {
	var b strings.Builder
	// Порядок обхода мапы случаен — ключи сортируем явно.
	for _, k := range slices.Sorted(maps.Keys(params)) {
		fmt.Fprintf(&b, "%s=%s&", k, params[k])
	}
	return strings.TrimSuffix(b.String(), "&")
}

func main() {
	params := map[string]string{"q": "go", "page": "2", "lang": "ru", "sort": "new"}
	for range 3 {
		fmt.Println(query(params))
	}
}
