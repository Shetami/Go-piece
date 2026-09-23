package main

import "fmt"

// removeEmpty удаляет пустые строки, сохраняя порядок остальных.
func removeEmpty(items []string) []string {
	for i := 0; i < len(items); i++ {
		if items[i] == "" {
			items = append(items[:i], items[i+1:]...)
		}
	}
	return items
}

func main() {
	fmt.Printf("%q\n", removeEmpty([]string{"a", "", "", "b", "", "c"}))
}
