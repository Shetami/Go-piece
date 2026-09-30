package main

import (
	"errors"
	"runtime"
)

// StackError — ошибка со стеком вызовов, снятым там, где её впервые
// обернули.
type StackError struct {
	Err    error
	Frames []string // полные имена функций: Frames[0] — тот, кто вызвал WithStack
}

// Error возвращает ровно текст Err: стек — для лога, не для сообщения.
func (e *StackError) Error() string { return e.Err.Error() }

// Unwrap открывает Err для errors.Is и errors.As.
func (e *StackError) Unwrap() error { return e.Err }

// WithStack оборачивает err, запоминая стек вызывающего (не больше 32 кадров).
//   - nil → nil.
//   - Если в цепочке err уже есть *StackError, стек второй раз не снимают и
//     возвращают err как есть: самый глубокий стек — самый ценный.
//   - Frames[0] — функция, вызвавшая WithStack; ни WithStack, ни
//     runtime.Callers в стеке быть не должно.
func WithStack(err error) error {
	if err == nil {
		return nil // именно нетипизированный nil, а не (*StackError)(nil)
	}
	var se *StackError
	if errors.As(err, &se) {
		return err
	}
	pcs := make([]uintptr, 32)
	// skip=2: пропустить сам runtime.Callers и WithStack.
	n := runtime.Callers(2, pcs)
	// CallersFrames правильно разворачивает встроенные (inlined) функции;
	// runtime.FuncForPC по каждому pc их бы потерял.
	frames := runtime.CallersFrames(pcs[:n])
	var names []string
	for {
		f, more := frames.Next()
		names = append(names, f.Function)
		if !more {
			break
		}
	}
	return &StackError{Err: err, Frames: names}
}

// StackOf возвращает Frames первого *StackError в цепочке err или nil.
func StackOf(err error) []string {
	var se *StackError
	if errors.As(err, &se) {
		return se.Frames
	}
	return nil
}
