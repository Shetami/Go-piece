package main

import (
	"errors"
	"fmt"
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
	// ваш код
}

// NewCache: ttl — сколько значение свежее, negTTL — сколько помнить
// «не найдено», now — часы.
func NewCache(load func(key string) (string, error), ttl, negTTL time.Duration, now func() time.Time) *Cache {
	// ваш код
	return &Cache{}
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
	// ваш код
	return "", nil
}
