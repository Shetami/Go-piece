package main

import "fmt"

type Worker struct {
	jobs chan string
	done chan struct{}
}

func (w *Worker) Start() {
	go func() {
		defer close(w.done)
		for j := range w.jobs {
			fmt.Println("обработал", j)
		}
	}()
}

func (w *Worker) Stop() {
	close(w.jobs)
	<-w.done
}

func main() {
	w := &Worker{done: make(chan struct{})}
	w.Start()
	w.jobs <- "письмо"
	w.Stop()
	fmt.Println("готово")
}
