package main

import "fmt"

func main() {
	go func() {
		for i := 1; i <= 3; i++ {
			fmt.Println("шаг", i)
		}
	}()
	fmt.Println("main закончил")
}
