package main

import "fmt"

type T struct{ n int }

func (t T) Val()  { fmt.Println("значение:", t.n) }
func (t *T) Ptr() { fmt.Println("указатель:", t.n) }

func main() {
	t := T{n: 1}
	defer t.Val()
	defer t.Ptr()

	t.n = 2
}
