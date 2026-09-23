package main

import "fmt"

type Stack[T any] struct {
	items []T
}

// Push меняет стек, поэтому получатель — указатель.
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Len() int { return len(s.items) }

func main() {
	var s Stack[string]
	s.Push("a")
	s.Push("b")
	fmt.Println("в стеке:", s.Len())
}
