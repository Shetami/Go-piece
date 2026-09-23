package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func (u User) String() string {
	// Печатаем поля, а не саму структуру: %v от u снова вызвал бы String.
	return fmt.Sprintf("User(%s, %d)", u.Name, u.Age)
}

func main() {
	fmt.Println(User{Name: "Аня", Age: 30})
}
