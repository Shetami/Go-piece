package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("не найдено")

func findUser(id int) error { return ErrNotFound }

func loadProfile(id int) error {
	if err := findUser(id); err != nil {
		// %w, а не %v: иначе связь с ErrNotFound теряется и errors.Is её не найдёт.
		return fmt.Errorf("профиль %d: %w", id, err)
	}
	return nil
}

func main() {
	err := loadProfile(7)
	if errors.Is(err, ErrNotFound) {
		fmt.Println("404:", err)
	} else {
		fmt.Println("500:", err)
	}
}
