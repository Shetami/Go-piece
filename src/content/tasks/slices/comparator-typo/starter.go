package main

import (
	"fmt"
	"sort"
)

type person struct {
	name string
	age  int
}

func main() {
	people := []person{{"Вика", 35}, {"Аня", 20}, {"Боря", 28}}

	sort.Slice(people, func(i, j int) bool {
		return people[i].age < people[i].age
	})

	fmt.Println(people)
}
