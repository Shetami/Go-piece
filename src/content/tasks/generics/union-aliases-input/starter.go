package main

import (
	"fmt"
	"maps"
	"slices"
)

type Set[T comparable] map[T]struct{}

func Of[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	for _, it := range items {
		s[it] = struct{}{}
	}
	return s
}

// Union возвращает новое множество — объединение a и b.
func Union[T comparable](a, b Set[T]) Set[T] {
	out := a
	for k := range b {
		out[k] = struct{}{}
	}
	return out
}

func main() {
	admins := Of("аня")
	editors := Of("боря", "вика")

	staff := Union(admins, editors)

	fmt.Println("админы:", slices.Sorted(maps.Keys(admins)))
	fmt.Println("все:", slices.Sorted(maps.Keys(staff)))
}
