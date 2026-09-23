package main

import "fmt"

// preview обрезает заголовок до n символов и ставит многоточие.
func preview(title string, n int) string {
	// Считаем и режем руны, а не байты: кириллица занимает по два байта.
	runes := []rune(title)
	if len(runes) <= n {
		return title
	}
	return string(runes[:n]) + "…"
}

func main() {
	fmt.Printf("%q\n", preview("Hello, world", 5))
	fmt.Printf("%q\n", preview("Привет, мир", 5))
}
