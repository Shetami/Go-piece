package main

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// ErrGoexit — f завершила горутину через runtime.Goexit, не вернув результат.
var ErrGoexit = errors.New("goroutine exited via runtime.Goexit")

// PanicError — паника внутри f, превращённая в ошибку.
type PanicError struct {
	Value any    // значение, переданное в panic
	Stack []byte // стек горутины В МОМЕНТ ПАНИКИ (с функцией, которая паникует)
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Value)
}

// Unwrap возвращает Value, если это error, иначе nil.
func (e *PanicError) Unwrap() error {
	err, _ := e.Value.(error)
	return err
}

// Go запускает f в новой горутине и возвращает канал, в который придёт
// ровно одно значение: ошибка f (или nil), *PanicError при панике, или
// ErrGoexit, если f вызвала runtime.Goexit. Паника в f не роняет процесс.
// Если результат никто не прочитает, горутина всё равно завершится.
func Go(f func() error) <-chan error {
	// Буфер 1: отправка не блокируется, даже если результат никому не нужен.
	ch := make(chan error, 1)
	go func() {
		returned := false
		defer func() {
			if returned {
				return
			}
			// f не вернулась: либо паника, либо Goexit (recover вернёт nil).
			// С Go 1.21 panic(nil) даёт *runtime.PanicNilError, не nil.
			if r := recover(); r != nil {
				// Стек снимаем здесь: отложенная функция выполняется
				// поверх кадров паникующей горутины, и они ещё видны.
				ch <- &PanicError{Value: r, Stack: debug.Stack()}
				return
			}
			ch <- ErrGoexit
		}()
		err := f()
		returned = true
		ch <- err
	}()
	return ch
}
