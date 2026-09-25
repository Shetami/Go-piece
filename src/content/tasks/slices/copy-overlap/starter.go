package main

import "fmt"

func main() {
	s := []int{1, 2, 3, 4, 5}
	n := copy(s[1:], s)
	fmt.Println(n, s)

	t := []int{1, 2, 3, 4, 5}
	n = copy(t, t[2:])
	fmt.Println(n, t)
}
