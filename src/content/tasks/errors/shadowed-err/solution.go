package main

import (
	"errors"
	"fmt"
	"strconv"
)

// parseAll разбирает все строки и сообщает о первой неудаче.
func parseAll(inputs []string) ([]int, error) {
	var out []int
	var err error

	for _, s := range inputs {
		// Присваивание, а не объявление: n объявляем отдельно,
		// чтобы := не завело внутри цикла новый err.
		var n int
		n, err = strconv.Atoi(s)
		if err != nil {
			continue
		}
		out = append(out, n)
	}

	if err != nil {
		return nil, fmt.Errorf("разбор: %w", err)
	}
	return out, nil
}

func main() {
	nums, err := parseAll([]string{"1", "два", "3"})

	fmt.Println(nums, err)
	fmt.Println(errors.Is(err, strconv.ErrSyntax))
}
