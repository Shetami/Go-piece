package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func findUser(id int) (string, error) {
	if id != 1 {
		return "", fmt.Errorf("user %d: %w", id, ErrNotFound)
	}
	return "аня", nil
}

func status(id int) int {
	_, err := findUser(id)
	// errors.Is проходит по всей цепочке обёрток.
	if errors.Is(err, ErrNotFound) {
		return 404
	}
	if err != nil {
		return 500
	}
	return 200
}

func main() {
	fmt.Println(status(1), status(42))
}
