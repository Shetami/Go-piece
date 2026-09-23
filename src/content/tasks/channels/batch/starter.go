package main

import "time"

// Batch собирает значения из in в пачки. Пачка отправляется, когда
// в ней size элементов или с первого элемента пачки прошло maxWait.
// Когда in закрылся, недобранная пачка отправляется, и выход закрывается.
// Пустые пачки не отправляются.
func Batch(in <-chan int, size int, maxWait time.Duration) <-chan []int {
	// ваш код
	return nil
}
