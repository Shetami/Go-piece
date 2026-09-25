package main

import (
	"fmt"
	"runtime"
	"strings"
)

// firstWord возвращает первое слово строки.
func firstWord(s string) string {
	i := strings.IndexByte(s, ' ')
	if i < 0 {
		return strings.Clone(s)
	}
	// Копия: подстрока без неё держала бы в памяти всю исходную строку.
	return strings.Clone(s[:i])
}

func main() {
	// 20 строк лога по мегабайту; сохраняем только первое слово каждой.
	var levels []string
	for range 20 {
		line := "ERROR " + strings.Repeat("x", 1<<20)
		levels = append(levels, firstWord(line))
	}

	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("слов: %d, первое: %s, в куче: %d МБ\n", len(levels), levels[0], ms.HeapAlloc>>20)
}
