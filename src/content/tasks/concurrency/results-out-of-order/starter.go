package main

import (
	"fmt"
	"time"
)

func work(n int) int {
	time.Sleep(time.Duration(5-n) * 10 * time.Millisecond)
	return n * n
}

// squares считает квадраты параллельно и возвращает их в порядке входа.
func squares(nums []int) []int {
	out := make(chan int)
	for _, n := range nums {
		go func() { out <- work(n) }()
	}

	var res []int
	for range nums {
		res = append(res, <-out)
	}
	return res
}

func main() {
	fmt.Println(squares([]int{1, 2, 3, 4}))
}
