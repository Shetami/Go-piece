package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"
)

func main() {
	votes := map[string]int{"go": 5, "rust": 3, "zig": 3, "python": 5, "java": 5, "c": 1, "ruby": 5, "kotlin": 3}

	names := slices.Collect(maps.Keys(votes))
	sort.Slice(names, func(i, j int) bool {
		return votes[names[i]] > votes[names[j]]
	})

	for i, name := range names {
		fmt.Printf("%d. %s %d\n", i+1, name, votes[name])
	}
}
