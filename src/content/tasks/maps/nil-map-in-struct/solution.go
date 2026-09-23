package main

import "fmt"

type Registry struct {
	items map[string]int
}

// NewRegistry — единственное место, где создаётся мапа: нулевой Registry
// не годится для записи.
func NewRegistry() *Registry {
	return &Registry{items: make(map[string]int)}
}

func (r *Registry) Add(name string) {
	r.items[name]++
}

func main() {
	r := NewRegistry()
	r.Add("go")
	r.Add("go")
	fmt.Println(r.items)
}
