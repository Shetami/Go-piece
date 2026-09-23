package main

import (
	"errors"
	"fmt"
)

func safe(f func()) (msg string) {
	defer func() {
		r := recover()
		msg = fmt.Sprintf("%T: %v", r, r)
	}()
	f()
	return "ok"
}

func main() {
	fmt.Println(safe(func() {}))
	fmt.Println(safe(func() { panic("строка") }))
	fmt.Println(safe(func() { panic(errors.New("ошибка")) }))
	fmt.Println(safe(func() {
		var m map[string]int
		m["a"] = 1
	}))
	fmt.Println(safe(func() {
		a := []int{}
		_ = a[5]
	}))
}
