package main

import "fmt"

func main() {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	close(ch)

	fmt.Println(len(ch), cap(ch))

	v, ok := <-ch
	fmt.Println(v, ok)
	v, ok = <-ch
	fmt.Println(v, ok)
	v, ok = <-ch
	fmt.Println(v, ok)

	fmt.Println(len(ch))
}
