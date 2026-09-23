package main

import (
	"errors"
	"fmt"
)

type StatusError interface {
	error
	Status() int
}

type notFound struct{ path string }

func (e notFound) Error() string { return e.path + " не найден" }
func (e notFound) Status() int   { return 404 }

func open(path string) error {
	return fmt.Errorf("открытие: %w", notFound{path: path})
}

func main() {
	err := open("/report")

	var se StatusError
	if errors.As(err, se) {
		fmt.Println("код", se.Status())
		return
	}
	fmt.Println("код 500")
}
