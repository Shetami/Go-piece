package main

import (
	"context"
	"slices"
	"sync"
)

// Fetcher загружает страницу url и возвращает ссылки с неё.
type Fetcher func(ctx context.Context, url string) ([]string, error)

// Crawl обходит страницы, начиная со start.
//
//   - Глубина start — 0, ссылки со страницы глубины d имеют глубину d+1.
//     Загружаются все страницы, до которых есть путь длиной не больше
//     maxDepth (считается по кратчайшему пути).
//   - Каждая страница загружается не больше одного раза.
//   - Одновременно выполняется не больше workers загрузок.
//   - Ошибка загрузки не останавливает обход: страница просто не попадает
//     в результат, и её ссылки не обходятся.
//   - Если ctx отменён, новые загрузки не начинаются, Crawl дожидается
//     начатых и возвращает то, что успел, и ctx.Err().
//
// Возвращает отсортированный список успешно загруженных страниц.
func Crawl(ctx context.Context, start string, maxDepth, workers int, fetch Fetcher) ([]string, error) {
	seen := map[string]bool{start: true}
	frontier := []string{start}
	var visited []string
	sem := make(chan struct{}, workers)

	// Обход по уровням: весь уровень d загружается раньше уровня d+1,
	// поэтому страница всегда обнаруживается на своей минимальной глубине.
	for depth := 0; depth <= maxDepth && len(frontier) > 0; depth++ {
		var (
			mu   sync.Mutex
			wg   sync.WaitGroup
			next []string
		)
	dispatch:
		for _, u := range frontier {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				break dispatch
			}
			if ctx.Err() != nil { // оба case могли быть готовы — select выбирает случайно
				<-sem
				break
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				links, err := fetch(ctx, u)
				if err != nil {
					return
				}
				mu.Lock()
				defer mu.Unlock()
				visited = append(visited, u)
				if depth == maxDepth {
					return
				}
				for _, l := range links {
					if !seen[l] { // seen трогаем только под mu
						seen[l] = true
						next = append(next, l)
					}
				}
			}()
		}
		wg.Wait() // не бросаем горутины даже при отмене
		if err := ctx.Err(); err != nil {
			slices.Sort(visited)
			return visited, err
		}
		frontier = next
	}
	slices.Sort(visited)
	return visited, nil
}
