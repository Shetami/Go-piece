package main

import (
	"fmt"
	"sort"
)

// median возвращает медиану, не трогая исходные данные.
func median(xs []int) int {
	sorted := xs
	sort.Ints(sorted)
	return sorted[len(sorted)/2]
}

func main() {
	temps := []int{30, 10, 20}
	fmt.Println("медиана:", median(temps))
	fmt.Println("исходные:", temps)
}
