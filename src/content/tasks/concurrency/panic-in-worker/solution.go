package main

import (
	"fmt"
	"sync"
)

func work(x int) int {
	if x == 3 {
		panic("плохие данные")
	}
	return x * x
}

func main() {
	results := make([]string, 5)

	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Паника в горутине роняет всю программу, если её не поймать
			// в этой же горутине: recover из main сюда не дотянется.
			defer func() {
				if r := recover(); r != nil {
					results[i] = fmt.Sprint("ошибка: ", r)
				}
			}()
			results[i] = fmt.Sprint(work(i))
		}()
	}
	wg.Wait()

	fmt.Println(results)
}
