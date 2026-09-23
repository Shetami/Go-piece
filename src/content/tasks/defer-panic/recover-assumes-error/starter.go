package main

import (
	"errors"
	"fmt"
)

func safely(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	f()
	return nil
}

func main() {
	fmt.Println(safely(func() { panic(errors.New("ошибка-значение")) }))
	fmt.Println(safely(func() { panic("просто строка") }))
	fmt.Println(safely(func() {
		var m map[string]int
		m["x"] = 1
	}))
}
