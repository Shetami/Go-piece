package main

import "fmt"

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return "поле " + e.Field + " заполнено неверно"
}

// validate возвращает *ValidationError, а не error.
func validate(name string) *ValidationError {
	if name == "" {
		return &ValidationError{Field: "name"}
	}
	return nil
}

func check(name string) error {
	return validate(name)
}

func main() {
	err := check("Гоша")

	fmt.Println(err)
	fmt.Println(err == nil)

	if err != nil {
		fmt.Println("сработала ветка обработки ошибки")
	}

	var plain error
	fmt.Println(plain == nil)
}
