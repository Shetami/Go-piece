package main

import "fmt"

func main() {
	arr := [3]int{1, 2, 3}
	var fromArr []int
	for i, v := range arr {
		if i == 0 {
			arr[1], arr[2] = 20, 30
		}
		fromArr = append(fromArr, v)
	}
	fmt.Println("массив:", fromArr)

	arr = [3]int{1, 2, 3}
	var fromPtr []int
	for i, v := range &arr {
		if i == 0 {
			arr[1], arr[2] = 20, 30
		}
		fromPtr = append(fromPtr, v)
	}
	fmt.Println("указатель:", fromPtr)

	sl := []int{1, 2, 3}
	var fromSlice []int
	for i, v := range sl {
		if i == 0 {
			sl[1], sl[2] = 20, 30
		}
		fromSlice = append(fromSlice, v)
	}
	fmt.Println("слайс:", fromSlice)
}
