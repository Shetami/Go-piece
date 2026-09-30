package main

import (
	"context"
	"time"
)

// Cache — конкурентный кэш, который сам загружает отсутствующие значения.
//
//   - Get возвращает свежее значение из кэша (now() раньше, чем момент
//     загрузки + ttl), иначе загружает его через load.
//   - Одновременные Get одного ключа вызывают load один раз и получают один
//     и тот же результат.
//   - Ошибки не кэшируются: следующий Get попробует загрузить снова.
//   - Загрузка одного ключа не блокирует Get других ключей.
//   - Если ctx вызывающего отменён, пока он ждёт загрузку, Get сразу
//     возвращает ctx.Err(), а загрузка продолжается и попадает в кэш.
type Cache[K comparable, V any] struct {
	ttl  time.Duration
	now  func() time.Time
	load func(K) (V, error)
	// ваши поля
}

func NewCache[K comparable, V any](ttl time.Duration, now func() time.Time, load func(K) (V, error)) *Cache[K, V] {
	return &Cache[K, V]{ttl: ttl, now: now, load: load}
}

func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	// ваш код
	var zero V
	return zero, nil
}
