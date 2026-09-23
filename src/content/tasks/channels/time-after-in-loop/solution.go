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
	// Таймер заводится один раз, до цикла, — иначе каждая итерация
	// начинала бы отсчёт заново.
	timeout := time.After(time.Second)
	for {
		select {
		case e, ok := <-events:
			if !ok {
				fmt.Println("источник закрылся")
				return
			}
			fmt.Println(e)
		case <-timeout:
			fmt.Println("общий таймаут")
			return
		}
	}
}
