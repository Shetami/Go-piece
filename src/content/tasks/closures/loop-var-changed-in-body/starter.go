package main

import "fmt"

func main() {
	var fs []func()
	for i := 0; i < 6; i++ {
		fs = append(fs, func() { fmt.Print(i, " ") })
		i++
	}

	for _, f := range fs {
		f()
	}
	fmt.Println()
}
