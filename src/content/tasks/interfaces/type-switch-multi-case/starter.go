package main

import (
	"fmt"
	"time"
)

type myErr struct{}

func (myErr) Error() string  { return "сломалось" }
func (myErr) String() string { return "myErr" }

func describe(v any) {
	switch x := v.(type) {
	case nil:
		fmt.Printf("nil: %T\n", x)
	case int, int64:
		fmt.Printf("целое: %T\n", x)
	case error:
		fmt.Printf("ошибка: %T\n", x)
	case fmt.Stringer:
		fmt.Printf("stringer: %T\n", x)
	default:
		fmt.Printf("другое: %T\n", x)
	}
}

func main() {
	describe(nil)
	describe(int64(7))
	describe(myErr{})
	describe(time.Second)
	describe((*myErr)(nil))
	describe(3.14)
}
