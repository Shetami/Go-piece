package main

import (
	"bytes"
	"fmt"
	"sync"
)

var pool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func greet(name string) string {
	buf := pool.Get().(*bytes.Buffer)
	// Из пула приходит объект в том состоянии, в каком его вернули, — очищаем.
	buf.Reset()
	defer pool.Put(buf)

	buf.WriteString("привет, ")
	buf.WriteString(name)
	return buf.String()
}

func main() {
	fmt.Println(greet("Аня"))
	fmt.Println(greet("Боря"))
}
