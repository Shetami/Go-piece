package main

import "fmt"

type counter struct{ n int }

func (c counter) report(start <-chan struct{}, out chan<- string) {
	<-start
	out <- fmt.Sprint("метод на значении: ", c.n)
}

func main() {
	start := make(chan struct{})
	r1, r2, r3 := make(chan string), make(chan string), make(chan string)

	x := 1
	c := counter{n: 1}

	go func(v int) {
		<-start
		r1 <- fmt.Sprint("аргумент: ", v)
	}(x)
	go func() {
		<-start
		r2 <- fmt.Sprint("замыкание: ", x)
	}()
	go c.report(start, r3)

	x = 2
	c.n = 2
	close(start)

	fmt.Println(<-r1)
	fmt.Println(<-r2)
	fmt.Println(<-r3)
}
