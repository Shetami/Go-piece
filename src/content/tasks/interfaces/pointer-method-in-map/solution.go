package main

import (
	"fmt"
	"maps"
	"slices"
)

type Counter struct {
	hits int
}

func (c *Counter) Inc() { c.hits++ }

func main() {
	// Значение в мапе не адресуемо, поэтому храним указатели — у них метод Inc есть.
	pages := map[string]*Counter{"/": {}, "/about": {}}

	for _, p := range []string{"/", "/about", "/"} {
		pages[p].Inc()
	}

	for _, p := range slices.Sorted(maps.Keys(pages)) {
		fmt.Println(p, pages[p].hits)
	}
}
