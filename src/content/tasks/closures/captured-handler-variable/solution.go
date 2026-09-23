package main

import "fmt"

func main() {
	routes := map[string]func(){}

	var handler func()
	for _, name := range []string{"главная", "профиль", "выход"} {
		handler = func() { fmt.Println("страница:", name) }
		// Кладём само значение функции, а не обёртку, которая прочитает
		// переменную handler когда-нибудь потом.
		routes[name] = handler
	}

	routes["главная"]()
	routes["профиль"]()
}
