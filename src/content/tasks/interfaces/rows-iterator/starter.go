package main

import "iter"

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
	// ваш код
	return func(yield func(T, error) bool) {}
}

// Collect собирает все элементы. Любая ошибка — nil и эта ошибка.
func Collect[T any](rows Rows[T]) ([]T, error) {
	// ваш код
	return nil, nil
}

// Filter возвращает Rows только с элементами, для которых keep(v) == true.
// Ошибку исходного итератора он отдаёт через Err, а сам реализует io.Closer
// и закрывает исходный, если тот — io.Closer.
func Filter[T any](rows Rows[T], keep func(T) bool) Rows[T] {
	// ваш код
	return rows
}
