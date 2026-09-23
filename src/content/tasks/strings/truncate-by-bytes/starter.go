package main

import "fmt"

// preview обрезает заголовок до n символов и ставит многоточие.
func preview(title string, n int) string {
	if len(title) <= n {
		return title
	}
	return title[:n] + "…"
}

func main() {
	fmt.Printf("%q\n", preview("Hello, world", 5))
	fmt.Printf("%q\n", preview("Привет, мир", 5))
}
