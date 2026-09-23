package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// startWatcher запускает наблюдателя и возвращает функцию его остановки.
// Кто запустил — тот и останавливает.
func startWatcher(parent context.Context) (stop func()) {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		<-ctx.Done()
	}()
	return cancel
}

func main() {
	var stops []func()
	for range 3 {
		stops = append(stops, startWatcher(context.Background()))
	}
	// Наблюдатели больше не нужны.
	for _, stop := range stops {
		stop()
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println("горутин:", runtime.NumGoroutine())
}
