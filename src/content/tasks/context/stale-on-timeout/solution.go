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

type flight[V any] struct {
	done chan struct{}
	val  V
	err  error
}

// Cache — кэш, который при медленной загрузке отдаёт устаревшее значение,
// а загрузку не бросает: она дописывает свежее значение в фоне.
type Cache[K comparable, V any] struct {
	ttl         time.Duration
	loadTimeout time.Duration
	load        func(context.Context, K) (V, error)
	now         func() time.Time

	mu       sync.Mutex
	data     map[K]entry[V]
	inflight map[K]*flight[V]
}

func NewCache[K comparable, V any](ttl, loadTimeout time.Duration, load func(context.Context, K) (V, error), now func() time.Time) *Cache[K, V] {
	return &Cache[K, V]{ttl: ttl, loadTimeout: loadTimeout, load: load, now: now,
		data: map[K]entry[V]{}, inflight: map[K]*flight[V]{}}
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
	c.mu.Lock()
	old, have := c.data[key]
	if have && c.now().Sub(old.at) < c.ttl {
		c.mu.Unlock()
		return old.val, false, nil
	}
	f, ok := c.inflight[key]
	if !ok {
		f = &flight[V]{done: make(chan struct{})}
		c.inflight[key] = f
		// Загрузка переживает вызывающего: иначе при таймаутах кэш никогда
		// не обновится, и мы будем вечно отдавать устаревшее.
		lctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), c.loadTimeout)
		go func() {
			defer cancel()
			v, err := c.load(lctx, key)
			c.mu.Lock()
			if err == nil {
				c.data[key] = entry[V]{v, c.now()}
			}
			delete(c.inflight, key)
			f.val, f.err = v, err
			c.mu.Unlock()
			close(f.done)
		}()
	}
	c.mu.Unlock()

	select {
	case <-f.done:
		if f.err == nil {
			return f.val, false, nil
		}
		if have {
			return old.val, true, nil // stale-if-error
		}
		var zero V
		return zero, false, f.err
	case <-ctx.Done():
		if have {
			return old.val, true, nil
		}
		var zero V
		return zero, false, context.Cause(ctx)
	}
}
