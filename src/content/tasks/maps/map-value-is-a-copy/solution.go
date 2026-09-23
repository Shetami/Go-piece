package main

import "fmt"

type counter struct{ hits int }

func main() {
	byValue := map[string]counter{"a": {}}
	c := byValue["a"]
	c.hits++
	fmt.Println(byValue["a"].hits)

	byValue["a"] = c
	fmt.Println(byValue["a"].hits)

	byPtr := map[string]*counter{"a": {}}
	p := byPtr["a"]
	p.hits++
	byPtr["a"].hits++
	fmt.Println(byPtr["a"].hits)
}
