package main

import "context"

// Future — результат задачи, которая выполняется в фоне.
type Future[T any] struct {
	// ваши поля
}

// Async запускает f в отдельной горутине и сразу возвращает Future.
// f выполняется ровно один раз. Паника в f не роняет программу, а
// становится ошибкой, в тексте которой есть слово "panic".
func Async[T any](f func() (T, error)) *Future[T] {
	// ваш код
	return &Future[T]{}
}

// Done возвращает канал, который закрывается, когда результат готов.
func (fu *Future[T]) Done() <-chan struct{} {
	// ваш код
	return nil
}

// Await ждёт результат или отмену ctx. Его можно вызывать сколько угодно
// раз из любых горутин — все получат один и тот же результат. Отмена ctx
// прерывает только ожидание: задача продолжает работать, и следующий
// Await получит её результат.
func (fu *Future[T]) Await(ctx context.Context) (T, error) {
	// ваш код
	var zero T
	return zero, nil
}

// Then возвращает Future[U], которое после завершения src вызывает g
// с его значением. Если src завершилось ошибкой, g не вызывается, а
// ошибка переходит в результат. Then не блокирует вызывающего.
func Then[T, U any](src *Future[T], g func(T) (U, error)) *Future[U] {
	// ваш код
	return &Future[U]{}
}
