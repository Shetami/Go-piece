package main

import (
	"fmt"
	"sync"
	"time"
)

type Queue struct {
	mu    sync.Mutex
	cond  *sync.Cond
	items []int
}

func NewQueue() *Queue {
	q := &Queue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Push(x int) {
	q.mu.Lock()
	q.items = append(q.items, x)
	q.mu.Unlock()
	q.cond.Broadcast()
}

// Pop ждёт, пока в очереди что-нибудь появится, и забирает первый элемент.
func (q *Queue) Pop() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		q.cond.Wait()
	}
	x := q.items[0]
	q.items = q.items[1:]
	return x
}

func main() {
	q := NewQueue()
	results := make(chan int)
	for range 3 {
		go func() { results <- q.Pop() }()
	}

	time.Sleep(10 * time.Millisecond) // все три потребителя уже ждут
	for i := 1; i <= 3; i++ {
		q.Push(i)
		time.Sleep(10 * time.Millisecond)
	}

	sum := 0
	for range 3 {
		sum += <-results
	}
	fmt.Println("сумма:", sum)
}
