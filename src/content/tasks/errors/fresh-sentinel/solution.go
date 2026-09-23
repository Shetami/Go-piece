package main

import (
	"errors"
	"fmt"
)

var ErrEmpty = errors.New("пустая корзина")

func checkout(items []string) error {
	if len(items) == 0 {
		// Возвращаем ту самую переменную: сравнение идёт по указателю, не по тексту.
		return ErrEmpty
	}
	return nil
}

func main() {
	err := checkout(nil)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrEmpty))
}
