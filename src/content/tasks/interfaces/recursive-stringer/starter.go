package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func (u User) String() string {
	return fmt.Sprintf("User(%v)", u)
}

func main() {
	fmt.Println(User{Name: "Аня", Age: 30})
}
