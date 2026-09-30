package main

import (
	"slices"
	"sync"
)

// ParallelSortFunc сортирует s на месте, стабильно по cmp: равные элементы
// сохраняют исходный порядок. s режется на не больше workers примерно равных
// кусков, каждый сортируется в своей горутине, потом куски сливаются.
// Одновременно работают не больше workers горутин, и к возврату все они
// завершены. Дополнительная память — один буфер длины len(s).
// cmp можно вызывать конкурентно. workers <= 0 — паника.
func ParallelSortFunc[T any](s []T, workers int, cmp func(a, b T) int) {
	if workers <= 0 {
		panic("ParallelSortFunc: workers должно быть больше нуля")
	}
	parts := min(workers, len(s))
	if parts <= 1 {
		slices.SortStableFunc(s, cmp)
		return
	}
	// bounds[i]..bounds[i+1] — i-й кусок. Куски не пересекаются, поэтому
	// горутины пишут в один массив без гонки.
	bounds := make([]int, parts+1)
	for i := range bounds {
		bounds[i] = i * len(s) / parts
	}
	var wg sync.WaitGroup
	for i := range parts {
		wg.Add(1)
		go func(part []T) {
			defer wg.Done()
			slices.SortStableFunc(part, cmp) // стабильная — иначе порядок равных потеряется
		}(s[bounds[i]:bounds[i+1]])
	}
	wg.Wait()

	// Сливаем соседние куски попарно, пока не останется один.
	buf := make([]T, len(s))
	for len(bounds) > 2 {
		next := bounds[:1]
		for i := 0; i+1 < len(bounds); i += 2 {
			if i+2 >= len(bounds) { // непарный последний кусок остаётся как есть
				next = append(next, bounds[i+1])
				break
			}
			lo, mid, hi := bounds[i], bounds[i+1], bounds[i+2]
			mergeStable(buf[lo:hi], s[lo:mid], s[mid:hi], cmp)
			copy(s[lo:hi], buf[lo:hi])
			next = append(next, hi)
		}
		bounds = next
	}
}

// mergeStable сливает a и b в dst; при равенстве берёт из a — это и есть стабильность.
func mergeStable[T any](dst, a, b []T, cmp func(x, y T) int) {
	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		if cmp(b[j], a[i]) < 0 {
			dst[k] = b[j]
			j++
		} else {
			dst[k] = a[i]
			i++
		}
		k++
	}
	k += copy(dst[k:], a[i:])
	copy(dst[k:], b[j:])
}
