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
	limiter := newLimiter(2)

	for _, user := range []string{"аня", "боря"} {
		allow := limiter
		for range 3 {
			fmt.Println(user, allow())
		}
	}
}
