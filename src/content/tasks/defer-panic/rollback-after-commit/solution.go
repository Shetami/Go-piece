package main

import (
	"errors"
	"fmt"
)

type Tx struct {
	done bool
}

func (t *Tx) Commit() error {
	t.done = true
	fmt.Println("commit")
	return nil
}

// Rollback откатывает транзакцию. Вызывать его после Commit безопасно.
func (t *Tx) Rollback() {
	// Транзакция уже завершена — откатывать нечего. Иначе defer Rollback
	// откатывал бы и успешные транзакции.
	if t.done {
		return
	}
	t.done = true
	fmt.Println("rollback")
}

func transfer(amount int) error {
	tx := &Tx{}
	defer tx.Rollback()

	if amount > 100 {
		return errors.New("недостаточно средств")
	}
	return tx.Commit()
}

func main() {
	fmt.Println(transfer(50))
	fmt.Println(transfer(500))
}
