package main

import "slices"

// SortParallel возвращает НОВЫЙ отсортированный слайс; s не меняется.
//
//   - Сортировка устойчивая: элементы, равные по cmp, сохраняют исходный
//     порядок.
//   - Половины сортируются параллельно (merge sort), но одновременно
//     работает не больше maxPar горутин, вызывающих cmp (вызывающая
//     горутина считается). maxPar < 1 считается как 1.
//   - Куски короче 16 элементов сортируются без новых горутин.
func SortParallel[T any](s []T, maxPar int, cmp func(a, b T) int) []T {
	out := slices.Clone(s)
	if len(out) < 2 {
		return out
	}
	buf := make([]T, len(out))
	// Слоты для ДОПОЛНИТЕЛЬНЫХ горутин: вызывающая уже работает.
	sem := make(chan struct{}, max(maxPar-1, 0))
	mergeSort(out, buf, sem, cmp)
	return out
}

func mergeSort[T any](s, buf []T, sem chan struct{}, cmp func(a, b T) int) {
	if len(s) < 16 {
		slices.SortStableFunc(s, cmp)
		return
	}
	mid := len(s) / 2
	select {
	case sem <- struct{}{}:
		// Есть свободный слот — левая половина в новой горутине.
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() { <-sem }()
			mergeSort(s[:mid], buf[:mid], sem, cmp)
		}()
		mergeSort(s[mid:], buf[mid:], sem, cmp)
		<-done
	default:
		// Слотов нет — не ждём (ждать в рекурсии = дедлок), делаем сами.
		mergeSort(s[:mid], buf[:mid], sem, cmp)
		mergeSort(s[mid:], buf[mid:], sem, cmp)
	}
	merge(s, buf, mid, cmp)
}

// merge сливает отсортированные s[:mid] и s[mid:] через buf.
func merge[T any](s, buf []T, mid int, cmp func(a, b T) int) {
	copy(buf, s)
	i, j, k := 0, mid, 0
	for i < mid && j < len(s) {
		// Строго меньше: при равенстве берём левый — это и есть устойчивость.
		if cmp(buf[j], buf[i]) < 0 {
			s[k] = buf[j]
			j++
		} else {
			s[k] = buf[i]
			i++
		}
		k++
	}
	k += copy(s[k:], buf[i:mid])
	copy(s[k:], buf[j:len(s)])
}
