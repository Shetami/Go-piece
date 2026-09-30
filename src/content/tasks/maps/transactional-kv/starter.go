package main

import "errors"

var ErrNoTx = errors.New("нет открытой транзакции")

// Store — хранилище ключ-значение с вложенными транзакциями.
//
//   - Set, Get, Delete — обычные; изменения сразу видны в Get и Count,
//     даже внутри незакоммиченной транзакции.
//   - Count(v) — сколько ключей сейчас имеют значение v. Должен работать
//     за O(1), без обхода ключей.
//   - Begin открывает транзакцию; они вкладываются.
//   - Rollback отменяет все изменения самой внутренней транзакции
//     (включая удаления) и закрывает её.
//   - Commit закрывает самую внутреннюю транзакцию, передавая её изменения
//     внешней: если потом откатить внешнюю, откатятся и они.
//     Commit без вложенности делает изменения постоянными.
//   - Rollback и Commit без открытой транзакции возвращают ErrNoTx.
//   - Стоимость Rollback — O(число ключей, изменённых в транзакции),
//     а не O(размер хранилища).
type Store struct {
	// ваши поля
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Set(k, v string) {
	// ваш код
}

func (s *Store) Get(k string) (string, bool) {
	// ваш код
	return "", false
}

func (s *Store) Delete(k string) {
	// ваш код
}

func (s *Store) Count(v string) int {
	// ваш код
	return 0
}

func (s *Store) Begin() {
	// ваш код
}

func (s *Store) Rollback() error {
	// ваш код
	return nil
}

func (s *Store) Commit() error {
	// ваш код
	return nil
}
