package main

import (
	"context"
	"iter"
)

// Values превращает канал в итератор: range по нему отдаёт значения из ch
// до закрытия канала, отмены ctx или break в теле цикла.
func Values[T any](ctx context.Context, ch <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			select {
			case v, ok := <-ch:
				if !ok {
					return
				}
				// yield вернул false — в цикле случился break или return.
				// Звать yield после этого нельзя: рантайм запаникует.
				if !yield(v) {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
}

// Stream запускает перебор seq в отдельной горутине и отдаёт значения в
// канал. Канал закрывается, когда seq кончился или отменён ctx. После
// отмены перебор seq прекращается (yield возвращает false), горутина
// завершается, даже если seq бесконечный и выход никто не читает.
func Stream[T any](ctx context.Context, seq iter.Seq[T]) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		// range по seq — это вызов seq с телом цикла в роли yield;
		// return из тела заставляет yield вернуть false, и seq обязан
		// остановиться.
		for v := range seq {
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
