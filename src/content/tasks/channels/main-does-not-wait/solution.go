package main

import "fmt"

func main() {
	done := make(chan struct{})
	go func() {
		// Закрытие done — сигнал «я всё»: main дождётся его перед выходом.
		defer close(done)
		for i := 1; i <= 3; i++ {
			fmt.Println("шаг", i)
		}
	}()
	<-done
	fmt.Println("main закончил")
}
