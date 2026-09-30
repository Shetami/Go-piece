package main

import "iter"

// Map лениво применяет f к каждому элементу seq: f вызывается только
// тогда, когда потребитель просит очередной элемент.
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return // break у потребителя — останавливаем и источник
			}
		}
	}
}

// Filter лениво оставляет элементы, для которых keep вернул true.
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

// Take отдаёт не больше n первых элементов seq. После n-го элемента
// из seq не запрашивается ни одного лишнего. Результат можно обходить
// повторно — каждый обход начинается заново.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return // даже не начинаем обход источника
		}
		i := 0 // счётчик внутри функции: у каждого обхода свой
		for v := range seq {
			if !yield(v) {
				return
			}
			i++
			if i == n {
				return // выходим сразу, не дожидаясь (n+1)-го элемента
			}
		}
	}
}
