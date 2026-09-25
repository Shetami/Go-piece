package main

import "fmt"

// trace печатает вход в функцию и возвращает функцию,
// которая напечатает выход.
func trace(name string) func() {
	fmt.Println("→", name)
	return func() { fmt.Println("←", name) }
}

func save() {
	defer trace("save")()
	fmt.Println("  сохраняю")
}

func load() {
	defer trace("load")
	fmt.Println("  загружаю")
}

func main() {
	save()
	load()
}
