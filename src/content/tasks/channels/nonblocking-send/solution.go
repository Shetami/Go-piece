package main

import "fmt"

// report отправляет метрику, но не должен задерживать основную работу:
// если очередь полна, метрика выбрасывается.
func report(metrics chan<- string, m string) bool {
	// select с default — неблокирующая отправка.
	select {
	case metrics <- m:
		return true
	default:
		return false
	}
}

func main() {
	metrics := make(chan string, 3)
	sent, dropped := 0, 0
	for i := 1; i <= 5; i++ {
		if report(metrics, fmt.Sprint("m", i)) {
			sent++
		} else {
			dropped++
		}
	}
	fmt.Println("отправлено:", sent, "выброшено:", dropped)
}
