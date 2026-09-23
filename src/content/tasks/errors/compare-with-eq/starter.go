package main

import (
	"errors"
	"fmt"
)

var ErrNoStock = errors.New("нет на складе")

func reserve(item string) error {
	return fmt.Errorf("резерв %q: %w", item, ErrNoStock)
}

func main() {
	err := reserve("книга")
	if err == ErrNoStock {
		fmt.Println("предложим аналог")
	} else if err != nil {
		fmt.Println("ошибка:", err)
	}
}
