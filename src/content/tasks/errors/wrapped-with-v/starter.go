package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("не найдено")

func findUser(id int) error { return ErrNotFound }

func loadProfile(id int) error {
	if err := findUser(id); err != nil {
		return fmt.Errorf("профиль %d: %v", id, err)
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
