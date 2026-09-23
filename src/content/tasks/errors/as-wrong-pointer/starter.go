package main

import (
	"errors"
	"fmt"
)

type TimeoutError struct{ Op string }

func (e TimeoutError) Error() string { return e.Op + ": таймаут" }

func read() error {
	return fmt.Errorf("чтение конфига: %w", TimeoutError{Op: "read"})
}

func main() {
	err := read()

	var te *TimeoutError
	if errors.As(err, &te) {
		fmt.Println("повторим:", te.Op)
	} else {
		fmt.Println("сдаёмся:", err)
	}
}
