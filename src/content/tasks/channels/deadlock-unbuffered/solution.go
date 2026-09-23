package main

import "fmt"

func main() {
	results := make(chan int)

	// Отправитель уезжает в свою горутину: небуферизованный канал держит
	// отправку до тех пор, пока кто-то не начнёт принимать, а принимать
	// в этой программе некому — приёмный цикл ниже по тексту.
	go func() {
		for i := 1; i <= 3; i++ {
			results <- i * i
		}
		close(results)
	}()

	sum := 0
	for v := range results {
		sum += v
	}

	fmt.Println("сумма квадратов:", sum)
}
