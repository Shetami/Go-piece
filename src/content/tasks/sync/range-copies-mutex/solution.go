package main

import (
	"fmt"
	"sync"
)

type Account struct {
	mu      sync.Mutex
	balance int
}

func main() {
	accounts := []Account{{balance: 100}, {balance: 200}}

	var wg sync.WaitGroup
	for i := range accounts {
		// Указатель на элемент слайса: range по значению дал бы копию счёта вместе с мьютексом.
		acc := &accounts[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			acc.mu.Lock()
			acc.balance += 10
			acc.mu.Unlock()
		}()
	}
	wg.Wait()

	for i := range accounts {
		fmt.Println(accounts[i].balance)
	}
}
