package main

import "fmt"

type Logger struct {
	prefix string
}

func (l *Logger) Log(msg string) {
	fmt.Println(l.prefix + msg)
}

type Service struct {
	*Logger
	name string
}

func NewService(name string) *Service {
	// Встроенный указатель надо создать: иначе Log вызовется на nil.
	return &Service{Logger: &Logger{prefix: "[" + name + "] "}, name: name}
}

func (s *Service) Run() {
	s.Log("старт " + s.name)
}

func main() {
	s := NewService("billing")
	s.Run()
}
