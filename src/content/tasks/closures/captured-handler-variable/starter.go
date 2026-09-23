package main

import "fmt"

func main() {
	routes := map[string]func(){}

	var handler func()
	for _, name := range []string{"главная", "профиль", "выход"} {
		handler = func() { fmt.Println("страница:", name) }
		routes[name] = func() { handler() }
	}

	routes["главная"]()
	routes["профиль"]()
}
