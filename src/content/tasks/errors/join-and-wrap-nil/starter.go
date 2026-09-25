package main

import (
	"errors"
	"fmt"
)

func main() {
	var none error

	fmt.Println(errors.Join(nil, none) == nil)
	fmt.Println(errors.Join(none, errors.New("диск полон")))

	wrapped := fmt.Errorf("загрузка: %w", none)
	fmt.Println(wrapped == nil)
	fmt.Println(wrapped)
	fmt.Println(errors.Unwrap(wrapped) == nil)
}
