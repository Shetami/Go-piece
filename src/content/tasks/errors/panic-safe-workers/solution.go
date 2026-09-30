package main

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

// PanicError — паника в задаче, превращённая в ошибку.
type PanicError struct {
	Task  int    // индекс задачи
	Value any    // значение, переданное в panic
	Stack []byte // стек горутины в момент паники (runtime/debug.Stack)
}

// Error: "задача <Task>: паника: <Value>".
func (e *PanicError) Error() string {
	return fmt.Sprintf("задача %d: паника: %v", e.Task, e.Value)
}

// Unwrap: если Value — ошибка (в том числе *runtime.PanicNilError после
// panic(nil) или runtime.Error после записи в nil-мапу), вернуть её, иначе nil.
func (e *PanicError) Unwrap() error {
	err, _ := e.Value.(error)
	return err
}

// RunAll запускает все задачи параллельно и ждёт их завершения.
// Результат — errors.Join ошибок в порядке задач (nil, если ошибок нет):
//   - обычная ошибка задачи i → "задача i: <ошибка>" с %w;
//   - паника в задаче i → *PanicError со стеком.
//
// Паника в одной задаче не должна ронять процесс и мешать остальным.
func RunAll(tasks ...func() error) error {
	errs := make([]error, len(tasks))
	var wg sync.WaitGroup
	for i, task := range tasks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// recover работает только в той горутине, где случилась паника:
			// defer в RunAll её бы не поймал, и процесс упал бы целиком.
			defer func() {
				// С Go 1.21 panic(nil) даёт *runtime.PanicNilError, так что
				// r != nil ловит и её.
				if r := recover(); r != nil {
					errs[i] = &PanicError{Task: i, Value: r, Stack: debug.Stack()}
				}
			}()
			if err := task(); err != nil {
				errs[i] = fmt.Errorf("задача %d: %w", i, err)
			}
		}()
	}
	wg.Wait()
	return errors.Join(errs...)
}
