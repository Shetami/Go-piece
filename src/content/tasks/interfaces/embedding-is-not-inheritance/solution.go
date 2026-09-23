package main

import "fmt"

type Animal struct{}

func (Animal) Sound() string   { return "..." }
func (a Animal) Speak() string { return "говорит " + a.Sound() }

type Dog struct{ Animal }

func (Dog) Sound() string { return "гав" }

type Speaker interface{ Speak() string }

func main() {
	d := Dog{}
	fmt.Println(d.Sound())
	fmt.Println(d.Speak())

	var s Speaker = d
	fmt.Println(s.Speak())
	fmt.Println(d.Animal.Sound())
}
