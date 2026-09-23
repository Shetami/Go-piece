package main

import (
	"fmt"
	"time"
)

func main() {
	events := make(chan string)
	go func() {
		for i := 1; i <= 5; i++ {
			time.Sleep(300 * time.Millisecond)
			events <- fmt.Sprint("событие ", i)
		}
		close(events)
	}()

	// Слушаем события не дольше секунды в сумме.
	for {
		select {
		case e, ok := <-events:
			if !ok {
				fmt.Println("источник закрылся")
				return
			}
			fmt.Println(e)
		case <-time.After(time.Second):
			fmt.Println("общий таймаут")
			return
		}
	}
}
