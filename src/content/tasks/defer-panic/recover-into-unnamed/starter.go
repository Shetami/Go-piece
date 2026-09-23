package main

import "fmt"

func safeDiv(a, b int) (int, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("деление: %v", r)
		}
	}()
	return a / b, err
}

func main() {
	fmt.Println(safeDiv(10, 2))
	fmt.Println(safeDiv(1, 0))
}
