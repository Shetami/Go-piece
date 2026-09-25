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
	for _, acc := range accounts {
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
