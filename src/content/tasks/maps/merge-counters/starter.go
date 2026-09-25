package main

import (
	"fmt"
	"maps"
)

// merge прибавляет счётчики из src к счётчикам в dst.
func merge(dst, src map[string]int) {
	maps.Copy(dst, src)
}

func main() {
	total := map[string]int{"go": 3, "rust": 1}
	merge(total, map[string]int{"go": 2, "zig": 4})
	merge(total, map[string]int{"rust": 1})
	fmt.Println(total)
}
