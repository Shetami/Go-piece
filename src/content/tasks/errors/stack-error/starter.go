package main

// StackError — ошибка со стеком вызовов, снятым там, где её впервые
// обернули.
type StackError struct {
	Err    error
	Frames []string // полные имена функций: Frames[0] — тот, кто вызвал WithStack
}

// Error возвращает ровно текст Err: стек — для лога, не для сообщения.
func (e *StackError) Error() string {
	// ваш код
	return ""
}

// Unwrap открывает Err для errors.Is и errors.As.
func (e *StackError) Unwrap() error {
	// ваш код
	return nil
}

// WithStack оборачивает err, запоминая стек вызывающего (не больше 32 кадров).
//   - nil → nil.
//   - Если в цепочке err уже есть *StackError, стек второй раз не снимают и
//     возвращают err как есть: самый глубокий стек — самый ценный.
//   - Frames[0] — функция, вызвавшая WithStack; ни WithStack, ни
//     runtime.Callers в стеке быть не должно.
func WithStack(err error) error {
	// ваш код
	return err
}

// StackOf возвращает Frames первого *StackError в цепочке err или nil.
func StackOf(err error) []string {
	// ваш код
	return nil
}
