package main

import (
	"errors"
	"sync"
)

var (
	ErrSameAccount  = errors.New("перевод самому себе")
	ErrNoAccount    = errors.New("нет такого счёта")
	ErrInsufficient = errors.New("недостаточно средств")
	ErrBadAmount    = errors.New("сумма должна быть положительной")
)

type Account struct {
	ID      int
	mu      sync.Mutex
	balance int
}

type Bank struct {
	accounts map[int]*Account
	audit    func(from, to, amount int) // вызывается после изменения балансов
}

func NewBank(balances map[int]int, audit func(from, to, amount int)) *Bank {
	b := &Bank{accounts: map[int]*Account{}, audit: audit}
	for id, bal := range balances {
		b.accounts[id] = &Account{ID: id, balance: bal}
	}
	return b
}

// Balance возвращает баланс счёта.
func (b *Bank) Balance(id int) int {
	a := b.accounts[id]
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

// Transfer переводит amount со счёта from на счёт to.
//   - amount <= 0 — ErrBadAmount; from == to — ErrSameAccount;
//     неизвестный счёт — ErrNoAccount; не хватает денег — ErrInsufficient.
//     В этих случаях балансы не меняются.
//   - Оба счёта блокируются на время перевода. Встречные переводы
//     A→B и B→A, идущие одновременно, не должны приводить к дедлоку.
//   - После изменения балансов, ещё под блокировками, вызывается
//     audit(from, to, amount). Если audit паникует — балансы
//     возвращаются к прежним, блокировки отпущены, а паника летит дальше.
func (b *Bank) Transfer(from, to, amount int) error {
	// ваш код
	return nil
}
