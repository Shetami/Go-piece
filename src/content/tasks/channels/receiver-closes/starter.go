package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)

	go func() {
		defer fmt.Println("производитель остановился")
		for i := 1; ; i++ {
			ch <- i
		}
	}()

	for v := range ch {
		if v == 3 {
			close(ch)
			break
		}
		fmt.Println(v)
	}

	time.Sleep(10 * time.Millisecond)
	fmt.Println("готово")
}
