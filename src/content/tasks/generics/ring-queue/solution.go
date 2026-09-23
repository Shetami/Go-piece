package main

// Queue — очередь FIFO на кольцевом буфере. Растёт по мере надобности.
type Queue[T any] struct {
	buf  []T
	head int // индекс самого старого элемента
	n    int // сколько элементов в очереди
}

func (q *Queue[T]) Push(v T) {
	if q.n == len(q.buf) {
		q.grow()
	}
	q.buf[(q.head+q.n)%len(q.buf)] = v
	q.n++
}

// Pop забирает самый старый элемент. На пустой очереди — нулевое значение и false.
func (q *Queue[T]) Pop() (T, bool) {
	var zero T
	if q.n == 0 {
		return zero, false
	}
	v := q.buf[q.head]
	q.buf[q.head] = zero // не держим ссылку: иначе сборщик мусора не освободит объект
	q.head = (q.head + 1) % len(q.buf)
	q.n--
	return v, true
}

func (q *Queue[T]) Len() int { return q.n }

// grow удваивает буфер и раскладывает элементы по порядку с нуля.
func (q *Queue[T]) grow() {
	buf := make([]T, max(4, 2*len(q.buf)))
	for i := range q.n {
		buf[i] = q.buf[(q.head+i)%len(q.buf)]
	}
	q.buf, q.head = buf, 0
}
