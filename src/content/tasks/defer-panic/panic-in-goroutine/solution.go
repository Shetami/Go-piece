package main

import (
	"fmt"
	"sync"
)

func process(id int) string {
	if id == 1 {
		panic("сбой в задаче 1")
	}
	return fmt.Sprint("задача ", id, ": ок")
}

// safeProcess перехватывает панику в той же горутине, где она случилась:
// recover в main до чужой горутины не дотягивается.
func safeProcess(id int) (res string) {
	defer func() {
		if r := recover(); r != nil {
			res = fmt.Sprint("задача ", id, ": ", r)
		}
	}()
	return process(id)
}

func main() {
	results := make([]string, 3)
	var wg sync.WaitGroup
	for i := range 3 {
		wg.Go(func() {
			results[i] = safeProcess(i)
		})
	}
	wg.Wait()

	for _, r := range results {
		fmt.Println(r)
	}
}
