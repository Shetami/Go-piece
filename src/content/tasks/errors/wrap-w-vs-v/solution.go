package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("не найдено")

func main() {
	w := fmt.Errorf("загрузка профиля: %w", ErrNotFound)
	v := fmt.Errorf("загрузка профиля: %v", ErrNotFound)

	fmt.Println(w)
	fmt.Println(v)
	fmt.Println(errors.Is(w, ErrNotFound), errors.Is(v, ErrNotFound))
	fmt.Println(w.Error() == v.Error())
	fmt.Println(errors.New("не найдено") == ErrNotFound)
}
