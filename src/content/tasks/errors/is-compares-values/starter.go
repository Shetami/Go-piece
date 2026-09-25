package main

import (
	"errors"
	"fmt"
)

type codeErr struct{ code int }

func (e codeErr) Error() string { return fmt.Sprintf("код %d", e.code) }

type ptrErr struct{ code int }

func (e *ptrErr) Error() string { return fmt.Sprintf("код %d", e.code) }

func main() {
	err := fmt.Errorf("запрос: %w", codeErr{404})
	fmt.Println(errors.Is(err, codeErr{404}))
	fmt.Println(errors.Is(err, codeErr{500}))

	perr := fmt.Errorf("запрос: %w", &ptrErr{404})
	fmt.Println(errors.Is(perr, &ptrErr{404}))

	fmt.Println(errors.New("нет") == errors.New("нет"))
}
