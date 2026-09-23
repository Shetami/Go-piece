package main

import "fmt"

func grow(s []int) {
	s[0] = 99
	s = append(s, 4)
	s[1] = 77
}

func main() {
	small := []int{1, 2, 3}
	grow(small)
	fmt.Println(small)

	roomy := make([]int, 3, 10)
	copy(roomy, []int{1, 2, 3})
	grow(roomy)
	fmt.Println(roomy, len(roomy))
	fmt.Println(roomy[:4])
}
