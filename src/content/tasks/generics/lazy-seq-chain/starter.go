package main

import "iter"

// Map лениво применяет f к каждому элементу seq: f вызывается только
// тогда, когда потребитель просит очередной элемент.
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	// ваш код
	return func(yield func(U) bool) {}
}

// Filter лениво оставляет элементы, для которых keep вернул true.
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	// ваш код
	return func(yield func(T) bool) {}
}

// Take отдаёт не больше n первых элементов seq. После n-го элемента
// из seq не запрашивается ни одного лишнего. Результат можно обходить
// повторно — каждый обход начинается заново.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	// ваш код
	return func(yield func(T) bool) {}
}
