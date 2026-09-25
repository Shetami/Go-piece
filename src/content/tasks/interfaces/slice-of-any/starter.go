package main

import "fmt"

func printAll(items []any) {
	for _, it := range items {
		fmt.Println(it)
	}
}

func main() {
	names := []string{"аня", "боря"}
	ids := []int{1, 2}

	printAll(names)
	printAll(ids)
}
