package main

import (
	"errors"
	"fmt"
	"sync"
)

// ProcessAll вызывает f для каждого элемента items, одновременно — не
// больше workers вызовов (workers >= 1), и возвращается, когда все вызовы
// закончились.
//
// Результат:
//   - nil, если все вызовы f успешны;
//   - иначе одна ошибка, объединяющая ВСЕ неудачи (errors.Join), в порядке
//     индексов элементов, а не в порядке, в котором вызовы завершились.
//     Каждая неудача оформлена как "item <индекс>: <текст ошибки f>" и
//     оборачивает исходную ошибку, чтобы errors.Is и errors.As её находили.
func ProcessAll(items []string, workers int, f func(string) error) error {
	// Слот на каждый элемент: воркеры пишут в разные индексы — без мьютекса
	// и без гонки, а порядок задан индексом, а не временем завершения.
	errs := make([]error, len(items))
	jobs := make(chan int)

	var wg sync.WaitGroup
	for range min(workers, len(items)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if err := f(items[i]); err != nil {
					errs[i] = fmt.Errorf("item %d: %w", i, err)
				}
			}
		}()
	}
	for i := range items {
		jobs <- i
	}
	close(jobs)
	wg.Wait() // после Wait все записи в errs видны этой горутине

	// Join пропускает nil и от пустого набора возвращает настоящий nil.
	return errors.Join(errs...)
}
