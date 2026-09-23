package main

import (
	"errors"
	"fmt"
)

// Cleanup копит функции очистки и выполняет их в обратном порядке.
type Cleanup struct {
	fns []func() error
}

// Add запоминает функцию очистки.
func (c *Cleanup) Add(f func() error) {
	c.fns = append(c.fns, f)
}

// Run выполняет все функции в порядке, обратном добавлению.
// Выполняются все, даже если какая-то вернула ошибку или запаниковала;
// ошибки и паники собираются в одну ошибку. Повторный Run ничего не делает.
func (c *Cleanup) Run() error {
	var errs []error
	for i := len(c.fns) - 1; i >= 0; i-- {
		errs = append(errs, runOne(c.fns[i]))
	}
	c.fns = nil
	return errors.Join(errs...)
}

// runOne вызывает одну функцию и превращает её панику в ошибку,
// чтобы остальные очистки всё равно выполнились.
func runOne(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("паника в очистке: %v", r)
		}
	}()
	return f()
}
