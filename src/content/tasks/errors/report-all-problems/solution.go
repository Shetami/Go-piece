package main

import (
	"errors"
	"fmt"
	"strings"
)

type User struct {
	Name  string
	Email string
	Age   int
}

// validate должна сообщить обо всех проблемах сразу, а не по одной.
func validate(u User) error {
	// Копим все ошибки; errors.Join отбросит nil и вернёт nil, если ошибок не было.
	var errs []error
	if u.Name == "" {
		errs = append(errs, errors.New("пустое имя"))
	}
	if !strings.Contains(u.Email, "@") {
		errs = append(errs, errors.New("email без @"))
	}
	if u.Age < 0 {
		errs = append(errs, errors.New("отрицательный возраст"))
	}
	return errors.Join(errs...)
}

func main() {
	fmt.Println(validate(User{Email: "anya.example", Age: -1}))
	fmt.Println(validate(User{Name: "Аня", Email: "a@b.c", Age: 30}))
}
