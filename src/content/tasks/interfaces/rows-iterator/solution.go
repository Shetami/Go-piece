package main

import (
	"io"
	"iter"
)

// Rows — итератор в стиле database/sql.Rows:
//
//	for rows.Next() { v := rows.Value() … }
//	if err := rows.Err(); err != nil { … }
//
// Реализация может дополнительно быть io.Closer — тогда её нужно закрыть.
type Rows[T any] interface {
	Next() bool
	Value() T
	Err() error
}

// All превращает Rows в iter.Seq2[T, error] для range-over-func:
//   - для каждого элемента — yield(v, nil);
//   - после последнего элемента, если rows.Err() != nil, — yield(нулевое T, err);
//   - если rows — io.Closer, Close вызывается ровно один раз, когда обход
//     закончился, в том числе при досрочном break; если ошибок чтения не
//     было, а Close вернул ошибку — она отдаётся последней парой (если цикл
//     не прерван);
//   - после того как yield вернул false, больше не вызывать ни yield, ни Next.
func All[T any](rows Rows[T]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		closed := false
		closeRows := func() error {
			if c, ok := rows.(io.Closer); ok && !closed {
				closed = true
				return c.Close()
			}
			return nil
		}
		// defer срабатывает и при break в цикле потребителя, и при панике в нём.
		defer closeRows()

		for rows.Next() {
			if !yield(rows.Value(), nil) {
				return
			}
		}
		var zero T
		if err := rows.Err(); err != nil {
			yield(zero, err)
			return
		}
		if err := closeRows(); err != nil {
			yield(zero, err)
		}
	}
}

// Collect собирает все элементы. Любая ошибка — nil и эта ошибка.
func Collect[T any](rows Rows[T]) ([]T, error) {
	var out []T
	for v, err := range All(rows) {
		if err != nil {
			return nil, err // break — All сам закроет rows
		}
		out = append(out, v)
	}
	return out, nil
}

type filtered[T any] struct {
	Rows[T] // Value и Err — от исходного итератора
	keep    func(T) bool
}

func (f *filtered[T]) Next() bool {
	for f.Rows.Next() {
		if f.keep(f.Rows.Value()) {
			return true
		}
	}
	return false
}

// Close пробрасывает закрытие: без него All не увидел бы io.Closer у исходного.
func (f *filtered[T]) Close() error {
	if c, ok := f.Rows.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Filter возвращает Rows только с элементами, для которых keep(v) == true.
// Ошибку исходного итератора он отдаёт через Err, а сам реализует io.Closer
// и закрывает исходный, если тот — io.Closer.
func Filter[T any](rows Rows[T], keep func(T) bool) Rows[T] {
	return &filtered[T]{Rows: rows, keep: keep}
}
