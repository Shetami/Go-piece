package main

import (
	"context"
	"iter"
)

// Values превращает канал в итератор: range по нему отдаёт значения из ch
// до закрытия канала, отмены ctx или break в теле цикла.
func Values[T any](ctx context.Context, ch <-chan T) iter.Seq[T] {
	// ваш код
	return func(yield func(T) bool) {}
}

// Stream запускает перебор seq в отдельной горутине и отдаёт значения в
// канал. Канал закрывается, когда seq кончился или отменён ctx. После
// отмены перебор seq прекращается (yield возвращает false), горутина
// завершается, даже если seq бесконечный и выход никто не читает.
func Stream[T any](ctx context.Context, seq iter.Seq[T]) <-chan T {
	// ваш код
	return nil
}
