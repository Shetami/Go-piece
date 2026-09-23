package main

import (
	"fmt"
	"unsafe"
)

type Bad struct {
	a bool
	b int64
	c bool
}

type Good struct {
	b int64
	a bool
	c bool
}

func main() {
	fmt.Println(unsafe.Sizeof(Bad{}), unsafe.Sizeof(Good{}))
	fmt.Println(unsafe.Alignof(int64(0)), unsafe.Offsetof(Bad{}.b))
	fmt.Println(unsafe.Sizeof(struct{}{}), unsafe.Sizeof([0]int64{}))
}
