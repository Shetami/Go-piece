package main

import (
	"fmt"
	"sync"
)

func main() {
	var once sync.Once
	calls := 0

	func() {
		defer func() { fmt.Println("recover:", recover()) }()
		once.Do(func() {
			calls++
			panic("сбой")
		})
	}()

	once.Do(func() {
		calls++
		fmt.Println("вторая попытка")
	})
	fmt.Println("вызовов:", calls)

	v := sync.OnceValue(func() int {
		calls++
		return calls * 10
	})
	fmt.Println(v(), v(), calls)
}
