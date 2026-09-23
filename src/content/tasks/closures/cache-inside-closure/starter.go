package main

import "fmt"

var calls int

func slowSquare(n int) int {
	calls++
	return n * n
}

// memo запоминает результаты f, чтобы не считать одно и то же дважды.
func memo(f func(int) int) func(int) int {
	return func(n int) int {
		cache := map[int]int{}
		if v, ok := cache[n]; ok {
			return v
		}
		v := f(n)
		cache[n] = v
		return v
	}
}

func main() {
	sq := memo(slowSquare)
	for _, n := range []int{4, 4, 4, 5, 5} {
		sq(n)
	}
	fmt.Println("вызовов:", calls)
}
