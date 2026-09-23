package main

import (
	"cmp"
	"fmt"
	"slices"
)

type person struct {
	name string
	age  int
}

func main() {
	people := []person{{"Вика", 35}, {"Аня", 20}, {"Боря", 28}}

	// slices.SortFunc получает сами элементы, а не индексы, —
	// перепутать i и j просто негде.
	slices.SortFunc(people, func(a, b person) int {
		return cmp.Compare(a.age, b.age)
	})

	fmt.Println(people)
}
