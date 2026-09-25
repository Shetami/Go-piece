package main

import (
	"context"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	stop1 := context.AfterFunc(ctx, func() { fmt.Println("первый") })
	fmt.Println("stop1:", stop1())

	done2 := make(chan struct{})
	stop2 := context.AfterFunc(ctx, func() {
		fmt.Println("второй")
		close(done2)
	})

	cancel()
	<-done2
	fmt.Println("stop2:", stop2())

	done3 := make(chan struct{})
	context.AfterFunc(ctx, func() {
		fmt.Println("третий")
		close(done3)
	})
	<-done3
}
