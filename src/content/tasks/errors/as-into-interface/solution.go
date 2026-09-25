package main

import (
	"errors"
	"fmt"
)

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
		// errors.As ищет по всей цепочке, и целью может быть интерфейс.
		var t interface{ Timeout() bool }
		if errors.As(err, &t) && t.Timeout() {
			fmt.Println(i, "таймаут — повторим")
		} else {
			fmt.Println(i, "фатально:", err)
		}
	}
}
