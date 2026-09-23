package main

import (
	"fmt"
	"sync"
)

type Account struct {
	mu      sync.Mutex
	id      int
	balance int
}

func transfer(from, to *Account, n int) {
	from.mu.Lock()
	defer from.mu.Unlock()
	to.mu.Lock()
	defer to.mu.Unlock()

	from.balance -= n
	to.balance += n
}

func main() {
	a := &Account{id: 1, balance: 100}
	b := &Account{id: 2, balance: 100}

	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() { transfer(a, b, 1) })
		wg.Go(func() { transfer(b, a, 1) })
	}
	wg.Wait()

	fmt.Println(a.balance, b.balance)
}
