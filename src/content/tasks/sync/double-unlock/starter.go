package main

import (
	"errors"
	"fmt"
	"sync"
)

type Account struct {
	mu      sync.Mutex
	balance int
}

func (a *Account) Withdraw(n int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.balance < n {
		a.mu.Unlock()
		return errors.New("недостаточно средств")
	}
	a.balance -= n
	return nil
}

func main() {
	a := &Account{balance: 100}
	fmt.Println(a.Withdraw(30))
	fmt.Println(a.Withdraw(100))
	fmt.Println("остаток:", a.balance)
}
