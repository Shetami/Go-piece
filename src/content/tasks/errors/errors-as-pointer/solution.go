package main

import (
	"errors"
	"fmt"
)

type HTTPError struct{ Code int }

func (e *HTTPError) Error() string { return fmt.Sprintf("http %d", e.Code) }

func fetch() error {
	return fmt.Errorf("fetch: %w", &HTTPError{Code: 404})
}

func main() {
	err := fetch()

	var he *HTTPError
	fmt.Println(errors.As(err, &he), he.Code)

	fmt.Println(errors.Is(err, &HTTPError{Code: 404}))

	var t interface{ Timeout() bool }
	fmt.Println(errors.As(err, &t))

	fmt.Println(err)
}
