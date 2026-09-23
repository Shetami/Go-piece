package main

import (
	"errors"
	"fmt"
)

var (
	ErrDisk = errors.New("нет диска")
	ErrNet  = errors.New("нет сети")
)

func main() {
	err := errors.Join(ErrDisk, nil, ErrNet)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrDisk), errors.Is(err, ErrNet))
	fmt.Println(errors.Unwrap(err) == nil)
	fmt.Println(errors.Join(nil, nil) == nil)

	multi := fmt.Errorf("%w; %w", ErrDisk, ErrNet)
	fmt.Println(multi, errors.Is(multi, ErrNet), errors.Unwrap(multi) == nil)
}
