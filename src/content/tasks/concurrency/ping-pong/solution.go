package main

import "fmt"

func main() {
	ping := make(chan string)
	pong := make(chan string)

	go func() {
		for msg := range ping {
			fmt.Println("получил", msg)
			pong <- msg + "!"
		}
		close(pong)
	}()

	for _, w := range []string{"a", "b"} {
		ping <- w
		fmt.Println("ответ", <-pong)
	}
	close(ping)

	_, ok := <-pong
	fmt.Println(ok)
}
