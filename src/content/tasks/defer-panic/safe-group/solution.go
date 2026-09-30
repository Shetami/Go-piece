package main

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

// PanicError — паника горутины, превращённая в ошибку.
type PanicError struct {
	Value any    // то, что вернул recover
	Stack []byte // стек горутины в момент паники (runtime/debug.Stack)
}

// Error содержит значение паники.
func (e *PanicError) Error() string {
	return fmt.Sprintf("паника: %v", e.Value)
}

// Unwrap возвращает Value, если это error, иначе nil, — чтобы errors.Is/As
// находили исходную ошибку, с которой паниковали.
func (e *PanicError) Unwrap() error {
	err, _ := e.Value.(error)
	return err
}

// Group запускает функции в отдельных горутинах и собирает их ошибки.
// Нулевое значение готово к работе.
type Group struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error
}

// Go запускает fn в новой горутине. Паника внутри fn не роняет программу,
// а становится *PanicError.
func (g *Group) Go(fn func() error) {
	g.wg.Add(1) // до go: иначе Wait может успеть увидеть нулевой счётчик
	go func() {
		defer g.wg.Done()
		if err := callSafe(fn); err != nil {
			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()
		}
	}()
}

// callSafe вызывает fn; recover стоит в той же горутине, где паника.
func callSafe(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Стек снимаем здесь: во время отложенного вызова кадры
			// паникующей функции ещё на месте.
			err = &PanicError{Value: r, Stack: debug.Stack()}
		}
	}()
	return fn()
}

// Wait ждёт завершения всех запущенных функций и возвращает все их ошибки
// и паники, объединённые через errors.Join. Если ошибок нет — nil.
func (g *Group) Wait() error {
	g.wg.Wait()
	g.mu.Lock()
	defer g.mu.Unlock()
	return errors.Join(g.errs...)
}
