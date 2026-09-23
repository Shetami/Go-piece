package main

import "fmt"

func main() {
	arr := [3]int{1, 2, 3}
	copyArr := arr
	view := arr[:]

	view[0] = 100
	copyArr[1] = 200

	fmt.Println(arr)
	fmt.Println(copyArr)
	fmt.Println(view)
}
