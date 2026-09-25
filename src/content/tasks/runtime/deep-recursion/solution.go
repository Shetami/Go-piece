package main

import (
	"fmt"
	"runtime/debug"
)

// sum возвращает 1 + 2 + ... + n.
// Цикл вместо рекурсии: глубина вызовов больше не зависит от n.
func sum(n int) int {
	total := 0
	for i := 1; i <= n; i++ {
		total += i
	}
	return total
}

func main() {
	// Как в тесном контейнере: стеку горутины разрешено не больше мегабайта.
	debug.SetMaxStack(1 << 20)
	fmt.Println(sum(1_000_000))
}
