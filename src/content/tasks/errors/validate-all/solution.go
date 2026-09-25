package main

import (
	"errors"
	"strings"
)

type User struct {
	Name  string
	Email string
	Age   int
}

// FieldError — ошибка в конкретном поле.
type FieldError struct {
	Field string
	Msg   string
}

func (e *FieldError) Error() string { return e.Field + ": " + e.Msg }

// Validate проверяет пользователя и возвращает ВСЕ найденные проблемы
// одной ошибкой (или nil). Каждая проблема — *FieldError:
//   - Name пустое        → Field "Name",  Msg "обязательно"
//   - Email без "@"      → Field "Email", Msg "нет @"
//   - Age < 0 или > 150  → Field "Age",   Msg "вне диапазона"
func Validate(u User) error {
	var errs []error
	if u.Name == "" {
		errs = append(errs, &FieldError{"Name", "обязательно"})
	}
	if !strings.Contains(u.Email, "@") {
		errs = append(errs, &FieldError{"Email", "нет @"})
	}
	if u.Age < 0 || u.Age > 150 {
		errs = append(errs, &FieldError{"Age", "вне диапазона"})
	}
	// Join от пустого списка — nil, так что успех отдельно проверять не нужно.
	return errors.Join(errs...)
}
