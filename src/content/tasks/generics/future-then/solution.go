package main

import (
	"context"
	"fmt"
)

// Future — результат задачи, которая выполняется в фоне.
type Future[T any] struct {
	done chan struct{} // закрывается, когда val и err записаны
	val  T
	err  error
}

// Async запускает f в отдельной горутине и сразу возвращает Future.
// f выполняется ровно один раз. Паника в f не роняет программу, а
// становится ошибкой, в тексте которой есть слово "panic".
func Async[T any](f func() (T, error)) *Future[T] {
	fu := &Future[T]{done: make(chan struct{})}
	go func() {
		// defer выполняются в обратном порядке: сначала recover
		// записывает ошибку, потом close публикует результат.
		defer close(fu.done)
		defer func() {
			if r := recover(); r != nil {
				fu.err = fmt.Errorf("panic: %v", r)
			}
		}()
		fu.val, fu.err = f()
	}()
	return fu
}

// Done возвращает канал, который закрывается, когда результат готов.
func (fu *Future[T]) Done() <-chan struct{} { return fu.done }

// Await ждёт результат или отмену ctx. Его можно вызывать сколько угодно
// раз из любых горутин — все получат один и тот же результат. Отмена ctx
// прерывает только ожидание: задача продолжает работать, и следующий
// Await получит её результат.
func (fu *Future[T]) Await(ctx context.Context) (T, error) {
	select {
	case <-fu.done:
		// close — событие для всех читателей сразу, а запись в val/err
		// произошла до него (happens-before), так что читать безопасно.
		return fu.val, fu.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// Then возвращает Future[U], которое после завершения src вызывает g
// с его значением. Если src завершилось ошибкой, g не вызывается, а
// ошибка переходит в результат. Then не блокирует вызывающего.
func Then[T, U any](src *Future[T], g func(T) (U, error)) *Future[U] {
	return Async(func() (U, error) {
		v, err := src.Await(context.Background())
		if err != nil {
			var zero U
			return zero, err
		}
		return g(v)
	})
}
