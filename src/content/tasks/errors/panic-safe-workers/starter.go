package main

// PanicError — паника в задаче, превращённая в ошибку.
type PanicError struct {
	Task  int    // индекс задачи
	Value any    // значение, переданное в panic
	Stack []byte // стек горутины в момент паники (runtime/debug.Stack)
}

// Error: "задача <Task>: паника: <Value>".
func (e *PanicError) Error() string {
	// ваш код
	return ""
}

// Unwrap: если Value — ошибка (в том числе *runtime.PanicNilError после
// panic(nil) или runtime.Error после записи в nil-мапу), вернуть её, иначе nil.
func (e *PanicError) Unwrap() error {
	// ваш код
	return nil
}

// RunAll запускает все задачи параллельно и ждёт их завершения.
// Результат — errors.Join ошибок в порядке задач (nil, если ошибок нет):
//   - обычная ошибка задачи i → "задача i: <ошибка>" с %w;
//   - паника в задаче i → *PanicError со стеком.
//
// Паника в одной задаче не должна ронять процесс и мешать остальным.
func RunAll(tasks ...func() error) error {
	// ваш код
	return nil
}
