package main

import (
	"fmt"
	"runtime"
)

func step() {
	defer fmt.Println("defer step")
	defer func() {
		fmt.Println("recover:", recover())
	}()

	runtime.Goexit()
	fmt.Println("после Goexit")
}

func main() {
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer fmt.Println("defer горутины")
		step()
		fmt.Println("после step")
	}()

	<-done
	fmt.Println("main дождался")
}
