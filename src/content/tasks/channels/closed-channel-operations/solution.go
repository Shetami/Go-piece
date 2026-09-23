package main

import "fmt"

func try(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name+":", r)
			return
		}
		fmt.Println(name + ": ok")
	}()
	f()
}

func main() {
	ch := make(chan int, 1)
	close(ch)

	try("чтение", func() { <-ch })
	try("отправка", func() { ch <- 1 })
	try("повторное закрытие", func() { close(ch) })

	var nilCh chan int
	try("закрытие nil", func() { close(nilCh) })
}
