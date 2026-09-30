package main

import (
	"slices"
	"sort"
)

// Interval — полуоткрытый промежуток [Start, End).
type Interval struct{ Start, End int }

// Set — множество целых чисел, которое хранится как отсортированный слайс
// непересекающихся и не касающихся интервалов. Нулевое значение — пустое множество.
type Set struct {
	iv []Interval
}

// Add добавляет [start, end). Пустой интервал (start >= end) игнорируется.
// Пересекающиеся и касающиеся интервалы склеиваются в один.
// Место ищется двоичным поиском.
func (s *Set) Add(start, end int) {
	if start >= end {
		return
	}
	// i — первый интервал, который заканчивается не раньше start (>=: касание тоже клеим).
	i := sort.Search(len(s.iv), func(k int) bool { return s.iv[k].End >= start })
	// j — первый интервал, который начинается строго правее end: с ним уже не склеить.
	j := sort.Search(len(s.iv), func(k int) bool { return s.iv[k].Start > end })
	if i < j { // s.iv[i:j] сливаются с новым
		start = min(start, s.iv[i].Start)
		end = max(end, s.iv[j-1].End)
	}
	// Replace заменяет s.iv[i:j] одним интервалом; при i == j это вставка.
	s.iv = slices.Replace(s.iv, i, j, Interval{start, end})
}

// Contains сообщает, лежит ли x в множестве. O(log n).
func (s *Set) Contains(x int) bool {
	i := sort.Search(len(s.iv), func(k int) bool { return s.iv[k].End > x })
	return i < len(s.iv) && s.iv[i].Start <= x
}

// Intervals возвращает копию текущих интервалов по возрастанию.
func (s *Set) Intervals() []Interval {
	return slices.Clone(s.iv)
}
