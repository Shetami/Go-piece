package main

import "fmt"

type Registry struct {
	items map[string]int
}

func (r *Registry) Add(name string) {
	r.items[name]++
}

func main() {
	r := &Registry{}
	r.Add("go")
	r.Add("go")
	fmt.Println(r.items)
}
