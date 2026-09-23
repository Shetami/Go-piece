package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// startWatcher запускает наблюдателя, который работает, пока жив контекст.
func startWatcher(parent context.Context) {
	ctx, _ := context.WithCancel(parent)
	go func() {
		<-ctx.Done()
	}()
}

func main() {
	for range 3 {
		startWatcher(context.Background())
	}
	// Наблюдатели больше не нужны.
	time.Sleep(10 * time.Millisecond)
	fmt.Println("горутин:", runtime.NumGoroutine())
}
