package main

import (
	"fmt"
	"unsafe"
)

func main() {
	s := "довольно длинная строка, чтобы было видно, что длина ни при чём"
	sl := make([]int, 1000)
	m := map[int]int{1: 1, 2: 2}
	var i any = 42
	ch := make(chan int, 100)
	f := func() {}
	arr := [3]int64{}

	fmt.Println(unsafe.Sizeof(s), unsafe.Sizeof(sl), unsafe.Sizeof(m))
	fmt.Println(unsafe.Sizeof(i), unsafe.Sizeof(ch), unsafe.Sizeof(f), unsafe.Sizeof(arr))
}
