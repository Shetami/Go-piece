package main

import (
	"context"
	"sync"
	"time"
)

type entry[V any] struct {
	val     V
	expires time.Time
}

// call — идущая загрузка; done закрывается, когда val и err записаны.
type call[V any] struct {
	done chan struct{}
	val  V
	err  error
}

// Cache — кэш с загрузкой при промахе и временем жизни записей.
//
//   - Свежая запись (загружена в момент t, now() < t+ttl) отдаётся из кэша.
//   - При промахе или истёкшей записи значение грузится через load.
//     Одновременные Get одного ключа вызывают load один раз и получают его
//     результат; Get других ключей загрузки не ждут.
//   - Ошибка load возвращается ждавшим, но не кэшируется.
//   - Каждый Get уважает свой ctx: при отмене возвращает ctx.Err(), не
//     дожидаясь загрузки. Отмена ctx одного вызывающего не срывает
//     загрузку для остальных.
type Cache[K comparable, V any] struct {
	ttl  time.Duration
	now  func() time.Time
	load func(ctx context.Context, key K) (V, error)

	mu       sync.Mutex
	entries  map[K]entry[V]
	inflight map[K]*call[V]
}

func NewCache[K comparable, V any](ttl time.Duration, now func() time.Time,
	load func(ctx context.Context, key K) (V, error)) *Cache[K, V] {
	return &Cache[K, V]{
		ttl: ttl, now: now, load: load,
		entries:  make(map[K]entry[V]),
		inflight: make(map[K]*call[V]),
	}
}

func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	c.mu.Lock()
	if e, ok := c.entries[key]; ok && c.now().Before(e.expires) {
		c.mu.Unlock()
		return e.val, nil
	}
	cl, ok := c.inflight[key]
	if !ok {
		cl = &call[V]{done: make(chan struct{})}
		c.inflight[key] = cl
		// Загрузка — в своей горутине и с контекстом, который не отменится
		// вместе с ctx первого вызывающего: результат нужен всем ждущим.
		go c.fill(context.WithoutCancel(ctx), key, cl)
	}
	c.mu.Unlock()

	select {
	case <-cl.done:
		return cl.val, cl.err
	case <-ctx.Done():
		var zero V
		return zero, ctx.Err()
	}
}

func (c *Cache[K, V]) fill(ctx context.Context, key K, cl *call[V]) {
	cl.val, cl.err = c.load(ctx, key)
	c.mu.Lock()
	if cl.err == nil { // ошибки не кэшируем — следующий Get попробует снова
		c.entries[key] = entry[V]{val: cl.val, expires: c.now().Add(c.ttl)}
	}
	delete(c.inflight, key)
	c.mu.Unlock()
	// Закрываем после записи в кэш: кто пришёл после done, увидит запись.
	close(cl.done)
}
