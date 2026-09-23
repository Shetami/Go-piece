package main

import "fmt"

type Greeter struct{ name string }

func (g Greeter) Hello() { fmt.Println("привет,", g.name) }
func (g *Greeter) Bye()  { fmt.Println("пока,", g.name) }

func main() {
	g := Greeter{name: "Аня"}

	hello := g.Hello
	bye := g.Bye
	later := func() { g.Hello() }

	g.name = "Боря"

	hello()
	bye()
	later()
}
