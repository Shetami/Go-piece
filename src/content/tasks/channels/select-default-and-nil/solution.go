package main

import "fmt"

func main() {
	var nilCh chan int
	ready := make(chan int, 1)

	select {
	case v := <-nilCh:
		fmt.Println("nil", v)
	default:
		fmt.Println("default")
	}

	ready <- 7
	select {
	case v := <-nilCh:
		fmt.Println("nil", v)
	case v := <-ready:
		fmt.Println("ready", v)
	}

	for _, v := range []int{1, 2} {
		select {
		case ready <- v:
			fmt.Println("отправили", v)
		default:
			fmt.Println("буфер полон", v)
		}
	}
}
