package main

import (
	"fmt"
	"maps"
	"slices"
)

// withAliases добавляет к каждому маршруту алиас с префиксом /v1.
func withAliases(routes map[string]string) {
	for path, handler := range routes {
		routes["/v1"+path] = handler
	}
}

func main() {
	routes := map[string]string{"/users": "users", "/orders": "orders", "/items": "items"}
	withAliases(routes)
	fmt.Println(len(routes))
	fmt.Println(slices.Sorted(maps.Keys(routes)))
}
