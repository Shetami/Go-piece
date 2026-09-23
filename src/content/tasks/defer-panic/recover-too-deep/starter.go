package main

import "fmt"

// safeDivide должна перехватывать панику от деления на ноль
// и возвращать её как ошибку.
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("паника: %v", r)
			}
		}()
	}()

	return a / b, nil
}

func main() {
	fmt.Println(safeDivide(10, 2))
	fmt.Println(safeDivide(10, 0))
}
