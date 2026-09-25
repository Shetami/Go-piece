package main

import "fmt"

func main() {
	// Сколько раз ездили по каждому маршруту.
	trips := [][]string{
		{"мск", "спб"},
		{"мск", "казань"},
		{"мск", "спб"},
	}

	counts := map[[]string]int{}
	for _, route := range trips {
		counts[route]++
	}

	fmt.Println(counts[[]string{"мск", "спб"}], counts[[]string{"мск", "казань"}], len(counts))
}
