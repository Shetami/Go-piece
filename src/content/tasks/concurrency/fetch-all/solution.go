package main

import (
	"errors"
	"fmt"
	"sync"
)

// FetchAll параллельно загружает все urls через fetch — по горутине на
// каждый уникальный URL. Повторяющиеся URL загружаются один раз.
//
// Возвращает карту url → тело для успешных загрузок (не nil, даже если
// успешных нет) и ошибку: errors.Join всех неудач в порядке первого
// появления URL во входе, каждая обёрнута как fmt.Errorf("%s: %w", url, err).
// Если неудач нет — nil.
func FetchAll(urls []string, fetch func(url string) (string, error)) (map[string]string, error) {
	// Дедупликация до запуска горутин: порядок uniq — порядок первого появления.
	var uniq []string
	seen := make(map[string]bool)
	for _, u := range urls {
		if !seen[u] {
			seen[u] = true
			uniq = append(uniq, u)
		}
	}

	bodies := make([]string, len(uniq))
	errs := make([]error, len(uniq))
	var wg sync.WaitGroup
	for i, u := range uniq {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Своя ячейка на горутину — мьютекс не нужен.
			bodies[i], errs[i] = fetch(u)
		}()
	}
	wg.Wait()

	res := make(map[string]string)
	var failed []error
	for i, u := range uniq {
		if errs[i] != nil {
			failed = append(failed, fmt.Errorf("%s: %w", u, errs[i]))
			continue
		}
		res[u] = bodies[i]
	}
	return res, errors.Join(failed...)
}
