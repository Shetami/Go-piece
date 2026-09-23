package main

import "fmt"

// newLimiter разрешает не больше max вызовов.
func newLimiter(max int) func() bool {
	count := 0
	return func() bool {
		count++
		return count <= max
	}
}

func main() {
	for _, user := range []string{"аня", "боря"} {
		// Свой лимитер на пользователя: каждый вызов newLimiter заводит новый count.
		allow := newLimiter(2)
		for range 3 {
			fmt.Println(user, allow())
		}
	}
}
