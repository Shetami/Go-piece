package main

import (
	"bytes"
	"fmt"
	"runtime"
)

// loadHeaders читает 20 файлов по мегабайту и оставляет от каждого
// только заголовок — первые 16 байт.
func loadHeaders() [][]byte {
	var headers [][]byte
	for range 20 {
		file := make([]byte, 1<<20)
		// Копия на 16 байт: подслайс держал бы весь мегабайтный массив.
		headers = append(headers, bytes.Clone(file[:16]))
	}
	return headers
}

func main() {
	h := loadHeaders()
	runtime.GC()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Println("заголовков:", len(h))
	fmt.Println("в куче больше 10 МБ:", ms.HeapAlloc > 10<<20)
	runtime.KeepAlive(h)
}
