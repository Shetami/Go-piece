package main

import "fmt"

type Queue struct {
	items []string
}

func (q *Queue) Push(s string) {
	q.items = append(q.items, s)
}

// LenFunc возвращает функцию, которая показывает текущую длину очереди.
// Её отдают в систему метрик один раз при старте.
func (q *Queue) LenFunc() func() int {
	// Замыкание держит указатель на очередь и читает поле при каждом вызове.
	return func() int { return len(q.items) }
}

func main() {
	q := &Queue{}
	gauge := q.LenFunc()

	q.Push("письмо")
	q.Push("счёт")
	fmt.Println("в очереди:", gauge())
}
