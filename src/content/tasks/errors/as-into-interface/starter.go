package main

import "fmt"

type netErr struct {
	timeout bool
}

func (e *netErr) Error() string { return "сеть недоступна" }
func (e *netErr) Timeout() bool { return e.timeout }

func fetch(i int) error {
	return fmt.Errorf("fetch %d: %w", i, &netErr{timeout: i%2 == 0})
}

func main() {
	for i := range 4 {
		err := fetch(i)
		if t, ok := err.(interface{ Timeout() bool }); ok && t.Timeout() {
			fmt.Println(i, "таймаут — повторим")
		} else {
			fmt.Println(i, "фатально:", err)
		}
	}
}
