package main

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
	// ваш код
	return nil
}
