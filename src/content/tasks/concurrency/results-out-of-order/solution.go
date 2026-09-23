package main

import (
	"fmt"
	"sync"
	"time"
)

func work(n int) int {
	time.Sleep(time.Duration(5-n) * 10 * time.Millisecond)
	return n * n
}

// squares считает квадраты параллельно и возвращает их в порядке входа.
func squares(nums []int) []int {
	// Каждая горутина пишет в свою ячейку: порядок задаёт индекс,
	// а не то, кто первым закончил.
	res := make([]int, len(nums))
	var wg sync.WaitGroup
	for i, n := range nums {
		wg.Go(func() { res[i] = work(n) })
	}
	wg.Wait()
	return res
}

func main() {
	fmt.Println(squares([]int{1, 2, 3, 4}))
}
