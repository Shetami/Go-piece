package main

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
	// ваш код
}

// Contains сообщает, лежит ли x в множестве. O(log n).
func (s *Set) Contains(x int) bool {
	// ваш код
	return false
}

// Intervals возвращает копию текущих интервалов по возрастанию.
func (s *Set) Intervals() []Interval {
	// ваш код
	return nil
}
