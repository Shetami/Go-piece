package main

import (
	"errors"
	"fmt"
)

func safely(f func()) (err error) {
	defer func() {
		// recover возвращает any: паниковать можно чем угодно, не только ошибкой.
		switch r := recover().(type) {
		case nil:
		case error:
			err = r
		default:
			err = fmt.Errorf("паника: %v", r)
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
