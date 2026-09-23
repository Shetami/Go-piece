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
	if u.Name == "" {
		return errors.New("пустое имя")
	}
	if !strings.Contains(u.Email, "@") {
		return errors.New("email без @")
	}
	if u.Age < 0 {
		return errors.New("отрицательный возраст")
	}
	return nil
}

func main() {
	fmt.Println(validate(User{Email: "anya.example", Age: -1}))
	fmt.Println(validate(User{Name: "Аня", Email: "a@b.c", Age: 30}))
}
