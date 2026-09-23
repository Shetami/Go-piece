package main

import (
	"errors"
	"fmt"
)

var ErrNoRows = errors.New("нет строк")

type QueryError struct {
	Query string
	Err   error
}

func (e *QueryError) Error() string { return e.Query + ": " + e.Err.Error() }

// Unwrap открывает вложенную ошибку для errors.Is и errors.As.
func (e *QueryError) Unwrap() error { return e.Err }

func findUser(id int) error {
	return &QueryError{Query: "SELECT user", Err: ErrNoRows}
}

func main() {
	err := findUser(1)
	if errors.Is(err, ErrNoRows) {
		fmt.Println("пользователь не найден")
	} else {
		fmt.Println("сбой базы:", err)
	}
}
