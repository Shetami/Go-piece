package main

import "fmt"

// safeDivide должна перехватывать панику от деления на ноль
// и возвращать её как ошибку.
func safeDivide(a, b int) (result int, err error) {
	// recover работает, только если его вызвала САМА отложенная функция.
	// Лишняя обёртка вокруг — и паника летит дальше.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("паника: %v", r)
		}
	}()

	return a / b, nil
}

func main() {
	fmt.Println(safeDivide(10, 2))
	fmt.Println(safeDivide(10, 0))
}
