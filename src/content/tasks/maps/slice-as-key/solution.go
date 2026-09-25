package main

import (
	"fmt"
	"strings"
)

// routeKey превращает маршрут в сравнимое значение: слайс ключом мапы быть не может.
func routeKey(route []string) string {
	return strings.Join(route, "→")
}

func main() {
	// Сколько раз ездили по каждому маршруту.
	trips := [][]string{
		{"мск", "спб"},
		{"мск", "казань"},
		{"мск", "спб"},
	}

	counts := map[string]int{}
	for _, route := range trips {
		counts[routeKey(route)]++
	}

	fmt.Println(counts[routeKey([]string{"мск", "спб"})], counts[routeKey([]string{"мск", "казань"})], len(counts))
}
