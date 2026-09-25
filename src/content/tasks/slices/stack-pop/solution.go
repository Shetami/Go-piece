package main

import "fmt"

type Stack struct {
	items []int
}

func (s *Stack) Push(x int) {
	s.items = append(s.items, x)
}

// Pop снимает верхний элемент. Для пустого стека возвращает false.
func (s *Stack) Pop() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	// Последний элемент — len-1: индексы идут с нуля.
	x := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return x, true
}

func main() {
	var s Stack
	for i := 1; i <= 3; i++ {
		s.Push(i)
	}
	for {
		x, ok := s.Pop()
		if !ok {
			break
		}
		fmt.Println(x)
	}
}
