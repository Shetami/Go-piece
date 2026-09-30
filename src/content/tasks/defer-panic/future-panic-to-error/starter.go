package main

import (
	"context"
	"fmt"
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
	// ваши поля
}

// Async запускает fn в новой горутине и сразу возвращает Future.
// Паника fn не роняет процесс: результатом становится *PanicError
// (значение и стек runtime/debug.Stack).
func Async[T any](fn func() (T, error)) *Future[T] {
	// ваш код
	return &Future[T]{}
}

// Get ждёт результата fn или отмены ctx.
//   - Get можно вызывать сколько угодно раз и из разных горутин одновременно:
//     все получают один и тот же результат, fn выполняется один раз.
//   - Если ctx отменён раньше — (нулевое значение, ctx.Err()); fn при этом
//     продолжает работать, и следующий Get может получить её результат.
func (f *Future[T]) Get(ctx context.Context) (T, error) {
	// ваш код
	var zero T
	return zero, nil
}
