package main

import (
	"fmt"
	"time"
)

func slowQuery() {
	start := time.Now()
	// Замыкание откладывает вычисление time.Since до выхода из функции.
	defer func() {
		fmt.Println("запрос занял", time.Since(start).Round(100*time.Millisecond))
	}()

	time.Sleep(200 * time.Millisecond)
}

func main() {
	slowQuery()
}
