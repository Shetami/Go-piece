package main

import (
	"errors"
	"fmt"
)

type NotFound struct{ Key string }

func (e NotFound) Error() string { return "нет ключа " + e.Key }

func (e NotFound) Is(target error) bool {
	_, ok := target.(NotFound)
	return ok
}

func main() {
	err := fmt.Errorf("кэш: %w", NotFound{Key: "user:1"})

	fmt.Println(errors.Is(err, NotFound{Key: "user:1"}))
	fmt.Println(errors.Is(err, NotFound{Key: "user:2"}))
	fmt.Println(errors.Is(err, NotFound{}))

	var nf NotFound
	fmt.Println(errors.As(err, &nf), nf.Key)
}
