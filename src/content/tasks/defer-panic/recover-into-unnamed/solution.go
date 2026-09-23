package main

import "fmt"

// Результаты именованные: только так отложенная функция может
// записать ошибку в то, что вернётся вызывающему.
func safeDiv(a, b int) (res int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("деление: %v", r)
		}
	}()
	return a / b, nil
}

func main() {
	fmt.Println(safeDiv(10, 2))
	fmt.Println(safeDiv(1, 0))
}
