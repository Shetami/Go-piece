package main

import (
	"errors"
	"fmt"
)

var ErrTimeout = errors.New("таймаут")

func call(attempt int) error {
	if attempt < 3 {
		return fmt.Errorf("попытка %d: %w", attempt, ErrTimeout)
	}
	return nil
}

func main() {
	for attempt := 1; ; attempt++ {
		err := call(attempt)
		if err == nil {
			fmt.Println("успех с попытки", attempt)
			return
		}
		if errors.Is(ErrTimeout, err) {
			fmt.Println(err, "— повторяем")
			continue
		}
		fmt.Println("сдаёмся:", err)
		return
	}
}
