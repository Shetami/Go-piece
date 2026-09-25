package main

import (
	"fmt"
	"runtime/debug"
)

// sum возвращает 1 + 2 + ... + n.
func sum(n int) int {
	if n == 0 {
		return 0
	}
	return n + sum(n-1)
}

func main() {
	// Как в тесном контейнере: стеку горутины разрешено не больше мегабайта.
	debug.SetMaxStack(1 << 20)
	fmt.Println(sum(1_000_000))
}
