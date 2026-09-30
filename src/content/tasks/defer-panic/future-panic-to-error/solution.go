package main

import (
	"context"
	"fmt"
	"runtime/debug"
)

// PanicError — паника фоновой функции, превращённая в ошибку.
type PanicError struct {
	Value any
	Stack []byte
}

func (e *PanicError) Error() string { return fmt.Sprintf("паника в Async: %v", e.Value) }
func (e *PanicError) Unwrap() error { err, _ := e.Value.(error); return err }

// Future — результат функции, которая выполняется в фоне.
type Future[T any] struct {
	done chan struct{} // закрывается, когда val и err записаны
	val  T
	err  error
}

// Async запускает fn в новой горутине и сразу возвращает Future.
// Паника fn не роняет процесс: результатом становится *PanicError
// (значение и стек runtime/debug.Stack).
func Async[T any](fn func() (T, error)) *Future[T] {
	f := &Future[T]{done: make(chan struct{})}
	go func() {
		// close отложен первым — выполнится последним, когда результат
		// (в том числе после паники) уже записан.
		defer close(f.done)
		defer func() {
			if r := recover(); r != nil {
				var zero T
				f.val, f.err = zero, &PanicError{Value: r, Stack: debug.Stack()}
			}
		}()
		f.val, f.err = fn()
	}()
	return f
}

// Get ждёт результата fn или отмены ctx.
//   - Get можно вызывать сколько угодно раз и из разных горутин одновременно:
//     все получают один и тот же результат, fn выполняется один раз.
//   - Если ctx отменён раньше — (нулевое значение, ctx.Err()); fn при этом
//     продолжает работать, и следующий Get может получить её результат.
func (f *Future[T]) Get(ctx context.Context) (T, error) {
	select {
	case <-f.done:
		// Закрытый канал читается сколько угодно раз, а close даёт
		// happens-before: val и err видны без мьютекса.
		return f.val, f.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}
