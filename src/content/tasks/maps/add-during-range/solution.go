package main

import (
	"fmt"
	"maps"
	"slices"
)

// withAliases добавляет к каждому маршруту алиас с префиксом /v1.
func withAliases(routes map[string]string) {
	// Добавленные во время обхода ключи могут попасть в этот же обход,
	// поэтому алиасы собираем отдельно и добавляем после.
	aliases := make(map[string]string, len(routes))
	for path, handler := range routes {
		aliases["/v1"+path] = handler
	}
	maps.Copy(routes, aliases)
}

func main() {
	routes := map[string]string{"/users": "users", "/orders": "orders", "/items": "items"}
	withAliases(routes)
	fmt.Println(len(routes))
	fmt.Println(slices.Sorted(maps.Keys(routes)))
}
