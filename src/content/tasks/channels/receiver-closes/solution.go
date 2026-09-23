package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)
	// Отдельный канал «хватит»: его закрывает получатель,
	// а в ch по-прежнему пишет и отвечает за него только отправитель.
	done := make(chan struct{})

	go func() {
		defer fmt.Println("производитель остановился")
		for i := 1; ; i++ {
			select {
			case ch <- i:
			case <-done:
				return
			}
		}
	}()

	for v := range ch {
		if v == 3 {
			close(done)
			break
		}
		fmt.Println(v)
	}

	time.Sleep(10 * time.Millisecond)
	fmt.Println("готово")
}
