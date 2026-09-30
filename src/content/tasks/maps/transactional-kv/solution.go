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
	data   map[string]string
	counts map[string]int // значение → сколько ключей его имеют
	undo   []map[string]prevVal
}

// prevVal — состояние ключа до первого изменения в транзакции.
type prevVal struct {
	val     string
	existed bool
}

func NewStore() *Store {
	return &Store{data: make(map[string]string), counts: make(map[string]int)}
}

// remember записывает в журнал текущей транзакции старое состояние ключа —
// только при ПЕРВОМ изменении ключа в этой транзакции.
func (s *Store) remember(k string) {
	if len(s.undo) == 0 {
		return
	}
	log := s.undo[len(s.undo)-1]
	if _, ok := log[k]; ok {
		return
	}
	v, ok := s.data[k]
	log[k] = prevVal{v, ok}
}

// apply меняет данные и счётчики, без журнала.
func (s *Store) apply(k string, v string, present bool) {
	if old, ok := s.data[k]; ok {
		if s.counts[old]--; s.counts[old] == 0 {
			delete(s.counts, old)
		}
	}
	if present {
		s.data[k] = v
		s.counts[v]++
	} else {
		delete(s.data, k)
	}
}

func (s *Store) Set(k, v string) {
	s.remember(k)
	s.apply(k, v, true)
}

func (s *Store) Get(k string) (string, bool) {
	v, ok := s.data[k]
	return v, ok
}

func (s *Store) Delete(k string) {
	if _, ok := s.data[k]; !ok {
		return
	}
	s.remember(k)
	s.apply(k, "", false)
}

func (s *Store) Count(v string) int { return s.counts[v] }

func (s *Store) Begin() {
	s.undo = append(s.undo, make(map[string]prevVal))
}

func (s *Store) Rollback() error {
	if len(s.undo) == 0 {
		return ErrNoTx
	}
	log := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]
	for k, p := range log {
		s.apply(k, p.val, p.existed)
	}
	return nil
}

func (s *Store) Commit() error {
	if len(s.undo) == 0 {
		return ErrNoTx
	}
	log := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]
	if len(s.undo) == 0 {
		return nil // внешней нет — изменения уже в data, журнал не нужен
	}
	parent := s.undo[len(s.undo)-1]
	for k, p := range log {
		// У родителя остаётся более старое состояние, если оно уже записано.
		if _, ok := parent[k]; !ok {
			parent[k] = p
		}
	}
	return nil
}
