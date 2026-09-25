package main

import (
	"errors"
	"fmt"
)

type opErr struct {
	op  string
	err error
}

func (e *opErr) Error() string { return e.op + ": " + e.err.Error() }
func (e *opErr) Unwrap() error { return e.err }

func main() {
	base := errors.New("нет места")
	err := &opErr{"сохранение", fmt.Errorf("запись файла: %w", &opErr{"write", base})}
	fmt.Println(err)

	var target *opErr
	if errors.As(err, &target) {
		fmt.Println("нашёл:", target.op)
	}
	if errors.As(target.err, &target) {
		fmt.Println("глубже:", target.op)
	}

	fmt.Println(errors.Is(err, base))
}
