package main

import "fmt"

type Namer interface{ Name() string }

type User struct{ name string }

func (u User) Name() string { return u.name }

func main() {
	u := User{"Аня"}
	var n Namer = u
	u.name = "Боря"
	fmt.Println(n.Name(), u.Name())

	var np Namer = &u
	u.name = "Вика"
	fmt.Println(np.Name())
}
