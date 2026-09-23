package main

import "fmt"

// Count — итератор по числам от 0 до n-1.
func Count(n int) func(yield func(int) bool) {
	return func(yield func(int) bool) {
		for i := range n {
			// false от yield значит «цикл закончился» (break, return) — дальше нельзя.
			if !yield(i) {
				return
			}
		}
	}
}

func main() {
	for i := range Count(10) {
		if i == 3 {
			break
		}
		fmt.Println(i)
	}
	fmt.Println("готово")
}
