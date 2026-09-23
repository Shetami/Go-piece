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
		// Своё имя для ошибки итерации: := больше не затеняет внешний err,
		// а первая ошибка запоминается и не затирается следующими удачами.
		n, e := strconv.Atoi(s)
		if e != nil {
			if err == nil {
				err = e
			}
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
