package main

import "fmt"

// validate требует не меньше 8 символов.
func validate(pw string) string {
	if len(pw) < 8 {
		return "слишком короткий"
	}
	return "ок"
}

func main() {
	for _, pw := range []string{"secret12", "пароль", "пароль12", "qwerty"} {
		fmt.Printf("%s: %s\n", pw, validate(pw))
	}
}
