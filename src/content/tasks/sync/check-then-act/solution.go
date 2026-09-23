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
	time.Sleep(time.Millisecond) // проверка прав доступа
	// Проверка и захват — одной операцией: CAS записывает, только если
	// значение не изменилось с момента чтения, иначе пробуем заново.
	for {
		cur := active.Load()
		if cur >= limit {
			return false
		}
		if active.CompareAndSwap(cur, cur+1) {
			return true
		}
	}
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
