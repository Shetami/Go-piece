package main

import (
	"fmt"
	"sync"
)

// produce только пишет: канал общий, и закрыть его может лишь тот,
// кто знает, что писателей больше нет.
func produce(name string, n int, out chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := range n {
		out <- fmt.Sprintf("%s-%d", name, i)
	}
}

func main() {
	out := make(chan string)
	var wg sync.WaitGroup

	wg.Add(2)
	go produce("a", 3, out, &wg)
	go produce("b", 3, out, &wg)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(out)
		close(done)
	}()

	count := 0
	for range out {
		count++
	}
	<-done
	fmt.Println("получено:", count)
}
