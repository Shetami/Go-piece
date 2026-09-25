package main

import "fmt"

func handle(x int) int {
	return 100 / x
}

// safeHandle ловит панику одного элемента. recover работает на уровне
// функции, поэтому каждому элементу — свой вызов со своим defer.
func safeHandle(x int) (res int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return handle(x), nil
}

// processAll обрабатывает все элементы. Сбой одного
// не должен мешать остальным.
func processAll(items []int) {
	for _, x := range items {
		res, err := safeHandle(x)
		if err != nil {
			fmt.Println("сбой на", x, "-", err)
			continue
		}
		fmt.Println(x, "→", res)
	}
}

func main() {
	processAll([]int{5, 0, 20})
	fmt.Println("конец")
}
