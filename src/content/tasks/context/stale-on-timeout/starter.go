package main

import (
	"context"
	"sync"
	"time"
)

type entry[V any] struct {
	val V
	at  time.Time // когда загружено
}

// Cache — кэш, который при медленной загрузке отдаёт устаревшее значение,
// а загрузку не бросает: она дописывает свежее значение в фоне.
type Cache[K comparable, V any] struct {
	ttl         time.Duration
	loadTimeout time.Duration
	load        func(context.Context, K) (V, error)
	now         func() time.Time

	mu   sync.Mutex
	data map[K]entry[V]
	// ваши поля
}

func NewCache[K comparable, V any](ttl, loadTimeout time.Duration, load func(context.Context, K) (V, error), now func() time.Time) *Cache[K, V] {
	return &Cache[K, V]{ttl: ttl, loadTimeout: loadTimeout, load: load, now: now, data: map[K]entry[V]{}}
}

// Get возвращает значение по key.
//   - свежее значение (моложе ttl по часам now) — сразу, stale = false,
//     load не вызывается;
//   - иначе запускается загрузка (или используется уже идущая по этому
//     ключу — одновременно не больше одной). Загрузка идёт с контекстом,
//     который видит значения ctx, но не отменяется вместе с ним и ограничен
//     loadTimeout. Успешный результат попадает в кэш, даже если все
//     вызывающие ушли;
//   - загрузка успела — её результат, stale = false;
//   - ctx вызывающего кончился раньше, или загрузка вернула ошибку: если
//     есть устаревшее значение — оно, stale = true, err = nil; если нет —
//     context.Cause(ctx) или ошибка загрузки соответственно.
//
// Ошибки не кэшируются.
func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, bool, error) {
	// ваш код
	v, err := c.load(ctx, key)
	return v, false, err
}
