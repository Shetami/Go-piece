package main

// PQ — очередь с приоритетом на двоичной куче. Pop всегда снимает
// наименьший по less элемент. Push и Pop — O(log n).
// Пакет container/heap не использовать.
type PQ[T any] struct {
	items []T // куча: родитель i не больше детей 2i+1 и 2i+2
	less  func(a, b T) bool
}

func NewPQ[T any](less func(a, b T) bool) *PQ[T] {
	return &PQ[T]{less: less}
}

func (q *PQ[T]) Len() int { return len(q.items) }

func (q *PQ[T]) Push(v T) {
	q.items = append(q.items, v)
	// Всплытие: меняем с родителем, пока родитель больше.
	i := len(q.items) - 1
	for i > 0 {
		p := (i - 1) / 2
		if !q.less(q.items[i], q.items[p]) {
			break
		}
		q.items[i], q.items[p] = q.items[p], q.items[i]
		i = p
	}
}

// Pop снимает и возвращает наименьший элемент. Для пустой очереди —
// нулевое значение и false.
func (q *PQ[T]) Pop() (T, bool) {
	var zero T
	n := len(q.items)
	if n == 0 {
		return zero, false
	}
	top := q.items[0]
	q.items[0] = q.items[n-1]
	q.items[n-1] = zero // не держим ссылку на снятый элемент
	q.items = q.items[:n-1]

	// Погружение: меняем с меньшим из детей, пока он меньше нас.
	i, n := 0, n-1
	for {
		l, r, m := 2*i+1, 2*i+2, i
		if l < n && q.less(q.items[l], q.items[m]) {
			m = l
		}
		if r < n && q.less(q.items[r], q.items[m]) {
			m = r
		}
		if m == i {
			break
		}
		q.items[i], q.items[m] = q.items[m], q.items[i]
		i = m
	}
	return top, true
}
