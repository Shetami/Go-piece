package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const limit = 10

var active atomic.Int64

// tryAcquire занимает слот, если лимит ещё не исчерпан.
func tryAcquire() bool {
	if active.Load() < limit {
		time.Sleep(time.Millisecond) // проверка прав доступа
		active.Add(1)
		return true
	}
	return false
}

func main() {
	var granted atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			if tryAcquire() {
				granted.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("выдано слотов:", granted.Load(), "из", limit)
}
