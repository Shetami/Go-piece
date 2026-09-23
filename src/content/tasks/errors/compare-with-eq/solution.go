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
	// errors.Is идёт по цепочке обёрток; == сравнивает только верхнее звено.
	if errors.Is(err, ErrNoStock) {
		fmt.Println("предложим аналог")
	} else if err != nil {
		fmt.Println("ошибка:", err)
	}
}
