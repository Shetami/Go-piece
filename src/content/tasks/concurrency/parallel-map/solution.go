package main

import "sync"

// ParallelMap применяет f к каждому элементу и возвращает результаты
// В ТОМ ЖЕ ПОРЯДКЕ, что и вход.
//
// Одновременно должно работать не больше workers горутин: f может быть
// тяжёлой, и запускать её сразу на всём слайсе нельзя.
func ParallelMap(items []int, workers int, f func(int) int) []int {
	out := make([]int, len(items))
	if len(items) == 0 {
		return out
	}
	if workers < 1 {
		workers = 1
	}

	// Порядок держится сам собой: воркеры пишут в out[i] по индексу задачи.
	// Собирать результаты из канала и потом сортировать не нужно — и не надо,
	// потому что каждая ячейка достаётся ровно одной горутине, и гонки нет.
	tasks := make(chan int)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := range tasks {
				out[i] = f(items[i])
			}
		}()
	}

	for i := range items {
		tasks <- i
	}
	close(tasks)
	wg.Wait()

	return out
}
