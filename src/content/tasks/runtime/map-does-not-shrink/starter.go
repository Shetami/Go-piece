package main

import (
	"fmt"
	"runtime"
)

type entry [128]byte

func main() {
	cache := make(map[int]entry)
	for i := range 100_000 {
		cache[i] = entry{}
	}

	// Сбрасываем кэш: старые записи больше не нужны.
	clear(cache)

	for i := range 10 {
		cache[i] = entry{}
	}

	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Println("записей:", len(cache))
	fmt.Println("в куче больше 10 МБ:", ms.HeapAlloc > 10<<20)
	runtime.KeepAlive(cache)
}
