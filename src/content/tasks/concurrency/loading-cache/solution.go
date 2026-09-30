package main

import (
	"context"
	"sync"
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

	mu      sync.Mutex
	entries map[K]*entry[V]
}

type entry[V any] struct {
	ready   chan struct{} // закрыт — val/err/expires заполнены
	val     V
	err     error
	expires time.Time
}

func NewCache[K comparable, V any](ttl time.Duration, now func() time.Time, load func(K) (V, error)) *Cache[K, V] {
	return &Cache[K, V]{ttl: ttl, now: now, load: load, entries: make(map[K]*entry[V])}
}

func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	c.mu.Lock()
	e, ok := c.entries[key]
	if ok {
		select {
		case <-e.ready:
			if e.err == nil && c.now().Before(e.expires) {
				c.mu.Unlock()
				return e.val, nil
			}
			ok = false // протухло — загружаем заново
		default: // загрузка идёт — присоединяемся
		}
	}
	if !ok {
		e = &entry[V]{ready: make(chan struct{})}
		c.entries[key] = e
		// Загрузка в своей горутине: вызывающий может уйти по ctx,
		// а остальные всё равно получат результат.
		go c.fill(key, e)
	}
	c.mu.Unlock()

	select {
	case <-e.ready:
		return e.val, e.err
	case <-ctx.Done():
		var zero V
		return zero, ctx.Err()
	}
}

func (c *Cache[K, V]) fill(key K, e *entry[V]) {
	v, err := c.load(key) // без мьютекса: другие ключи не ждут
	c.mu.Lock()
	e.val, e.err = v, err
	e.expires = c.now().Add(c.ttl)
	if err != nil && c.entries[key] == e {
		delete(c.entries, key) // ошибку не кэшируем
	}
	c.mu.Unlock()
	close(e.ready)
}
