package main

import "fmt"

func run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("поймали: %v", r)
		}
	}()
	defer func() {
		panic("вторая")
	}()
	defer fmt.Println("отложенный вывод")

	panic("первая")
}

func main() {
	fmt.Println(run())
	fmt.Println("main дожил")
}
