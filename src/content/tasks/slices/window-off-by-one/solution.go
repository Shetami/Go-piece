package main

import "fmt"

// maxWindow возвращает наибольшую сумму k подряд идущих элементов.
func maxWindow(xs []int, k int) int {
	sum := 0
	for i := 0; i < k; i++ {
		sum += xs[i]
	}
	best := sum
	// Окно сдвигается до самого конца: последний элемент тоже должен в него войти.
	for i := k; i < len(xs); i++ {
		sum += xs[i] - xs[i-k]
		best = max(best, sum)
	}
	return best
}

func main() {
	fmt.Println(maxWindow([]int{1, 2, 3, 4, 10}, 2))
	fmt.Println(maxWindow([]int{5, 1, 1, 1}, 2))
}
