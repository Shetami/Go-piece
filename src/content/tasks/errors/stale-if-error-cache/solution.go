package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrNotFound — ключа нет в источнике. Это ответ, а не сбой: его кэшируют.
var ErrNotFound = errors.New("не найдено")

// StaleError — источник не ответил, но отдано устаревшее значение.
// Вызывающий может показать его пользователю и записать ошибку в лог.
type StaleError struct {
	Key string
	Age time.Duration // сколько значению на момент отдачи
	Err error         // почему не удалось обновить
}

func (e *StaleError) Error() string {
	return fmt.Sprintf("%s: устаревшее значение (%v): %v", e.Key, e.Age, e.Err)
}

func (e *StaleError) Unwrap() error { return e.Err }

// Cache — кэш перед медленным источником.
type Cache struct {
	load        func(string) (string, error)
	ttl, negTTL time.Duration
	now         func() time.Time

	mu      sync.Mutex
	entries map[string]entry
}

// NewCache: ttl — сколько значение свежее, negTTL — сколько помнить
// «не найдено», now — часы.
func NewCache(load func(key string) (string, error), ttl, negTTL time.Duration, now func() time.Time) *Cache {
	return &Cache{load: load, ttl: ttl, negTTL: negTTL, now: now, entries: make(map[string]entry)}
}

// Get возвращает значение по ключу:
//   - значение моложе ttl → (v, nil), load не вызывается;
//   - «не найдено» моложе negTTL → ("", ошибка с ErrNotFound), load не
//     вызывается;
//   - иначе вызвать load:
//   - успех → запомнить и вернуть (v, nil);
//   - ошибка с ErrNotFound → запомнить «не найдено», забыть старое
//     значение (его удалили — отдавать нельзя) и вернуть ("", ошибку load);
//   - другая ошибка → ничего не запоминать (следующий Get снова вызовет
//     load); если есть старое значение — вернуть (старое, *StaleError), иначе
//     ("", ошибку load).
//
// Безопасен для одновременного использования; load вызывается без
// удержания блокировки.
func (c *Cache) Get(key string) (string, error) {
	c.mu.Lock()
	e, ok := c.entries[key]
	now := c.now()
	if ok {
		age := now.Sub(e.at)
		if e.missing && age < c.negTTL {
			c.mu.Unlock()
			return "", e.err
		}
		if !e.missing && age < c.ttl {
			c.mu.Unlock()
			return e.val, nil
		}
	}
	c.mu.Unlock()

	// Источник медленный: держать мьютекс на время load — значит
	// заблокировать все ключи из-за одного.
	v, err := c.load(key)

	c.mu.Lock()
	defer c.mu.Unlock()
	now = c.now()
	switch {
	case err == nil:
		c.entries[key] = entry{val: v, at: now}
		return v, nil
	case errors.Is(err, ErrNotFound):
		// Негативный ответ кэшируем — иначе каждый запрос несуществующего
		// ключа бьёт в источник.
		c.entries[key] = entry{missing: true, err: err, at: now}
		return "", err
	}
	// Временную ошибку не кэшируем. Берём актуальную запись: её мог
	// обновить параллельный Get, пока мы ждали load.
	if e, ok := c.entries[key]; ok && !e.missing {
		return e.val, &StaleError{Key: key, Age: now.Sub(e.at), Err: err}
	}
	return "", err
}

type entry struct {
	val     string
	missing bool  // запомненное «не найдено»
	err     error // ошибка для missing
	at      time.Time
}
