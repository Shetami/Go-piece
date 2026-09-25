package main

import (
	"fmt"
	"sync"
)

func main() {
	var once sync.Once
	once.Do(func() { fmt.Println("первая") })
	once.Do(func() { fmt.Println("вторая") })

	calls := 0
	port := sync.OnceValue(func() int {
		calls++
		return 8080
	})
	fmt.Println(port(), port(), calls)
}
