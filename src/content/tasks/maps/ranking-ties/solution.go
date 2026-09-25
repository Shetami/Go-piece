package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	votes := map[string]int{"go": 5, "rust": 3, "zig": 3, "python": 5, "java": 5, "c": 1, "ruby": 5, "kotlin": 3}

	names := slices.Collect(maps.Keys(votes))
	slices.SortFunc(names, func(a, b string) int {
		// По голосам по убыванию, при равенстве — по имени: порядок из мапы случаен.
		return cmp.Or(
			cmp.Compare(votes[b], votes[a]),
			strings.Compare(a, b),
		)
	})

	for i, name := range names {
		fmt.Printf("%d. %s %d\n", i+1, name, votes[name])
	}
}
