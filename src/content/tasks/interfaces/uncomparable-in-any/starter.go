package main

import "fmt"

func equal(a, b any) (res bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return a == b, nil
}

func main() {
	fmt.Println(equal(1, 1))
	fmt.Println(equal(1, int64(1)))
	fmt.Println(equal([]int{1}, []int{1}))
	fmt.Println(equal([]int{1}, "x"))
	fmt.Println(equal(map[string]int{}, nil))
}
