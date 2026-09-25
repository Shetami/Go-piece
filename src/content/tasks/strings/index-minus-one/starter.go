package main

import (
	"fmt"
	"strings"
)

// domain возвращает домен из адреса почты в нижнем регистре.
// Для строки без @ возвращает пустую строку.
func domain(email string) string {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	host := email[at+1:]
	return strings.ToLower(host)
}

func main() {
	for _, e := range []string{"Anna@Example.COM", " bob@mail.ru ", "admin"} {
		fmt.Printf("%q\n", domain(e))
	}
}
