package main

import (
	"fmt"
	"runtime"
	"time"
)

func query(replica int) string {
	time.Sleep(time.Duration(replica) * 10 * time.Millisecond)
	return fmt.Sprintf("ответ реплики %d", replica)
}

// first опрашивает все реплики и возвращает самый быстрый ответ.
func first() string {
	ch := make(chan string)
	for r := 1; r <= 3; r++ {
		go func() { ch <- query(r) }()
	}
	return <-ch
}

func main() {
	for range 5 {
		fmt.Println(first())
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println("горутин:", runtime.NumGoroutine())
}
