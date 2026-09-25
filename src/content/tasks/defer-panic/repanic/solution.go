package main

import "fmt"

func work() {
	defer fmt.Println("defer 1")
	defer func() {
		r := recover()
		fmt.Println("work поймал:", r)
		panic(fmt.Sprint("снова: ", r))
	}()
	defer fmt.Println("defer 3")

	panic("бум")
}

func main() {
	defer func() {
		fmt.Println("main поймал:", recover())
		fmt.Println("ещё раз:", recover())
	}()

	work()
	fmt.Println("после work")
}
