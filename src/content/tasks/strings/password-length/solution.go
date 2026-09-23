package main

import (
	"fmt"
	"unicode/utf8"
)

// validate требует не меньше 8 символов.
func validate(pw string) string {
	// Символы, а не байты: кириллическая буква в UTF-8 весит два байта.
	if utf8.RuneCountInString(pw) < 8 {
		return "слишком короткий"
	}
	return "ок"
}

func main() {
	for _, pw := range []string{"secret12", "пароль", "пароль12", "qwerty"} {
		fmt.Printf("%s: %s\n", pw, validate(pw))
	}
}
