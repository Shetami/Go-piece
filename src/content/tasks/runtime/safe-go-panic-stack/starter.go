package main

import "errors"

// ErrGoexit — f завершила горутину через runtime.Goexit, не вернув результат.
var ErrGoexit = errors.New("goroutine exited via runtime.Goexit")

// PanicError — паника внутри f, превращённая в ошибку.
type PanicError struct {
	Value any    // значение, переданное в panic
	Stack []byte // стек горутины В МОМЕНТ ПАНИКИ (с функцией, которая паникует)
}

func (e *PanicError) Error() string {
	// ваш код
	return ""
}

// Unwrap возвращает Value, если это error, иначе nil.
func (e *PanicError) Unwrap() error {
	// ваш код
	return nil
}

// Go запускает f в новой горутине и возвращает канал, в который придёт
// ровно одно значение: ошибка f (или nil), *PanicError при панике, или
// ErrGoexit, если f вызвала runtime.Goexit. Паника в f не роняет процесс.
// Если результат никто не прочитает, горутина всё равно завершится.
func Go(f func() error) <-chan error {
	// ваш код
	ch := make(chan error, 1)
	close(ch)
	return ch
}
