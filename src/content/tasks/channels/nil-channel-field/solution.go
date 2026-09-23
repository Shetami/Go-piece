package main

import "fmt"

type Worker struct {
	jobs chan string
	done chan struct{}
}

// NewWorker создаёт оба канала: nil-канал в поле структуры блокирует навсегда.
func NewWorker() *Worker {
	return &Worker{jobs: make(chan string), done: make(chan struct{})}
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
	w := NewWorker()
	w.Start()
	w.jobs <- "письмо"
	w.Stop()
	fmt.Println("готово")
}
