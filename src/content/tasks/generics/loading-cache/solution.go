package main

import (
	"context"
	"sync"
	"time"
)

type cacheEntry[V any] struct {
	val V
	exp time.Time
}

// loadCall — загрузка, которая идёт прямо сейчас; её ждут все Get этого ключа.
type loadCall[V any] struct {
	done chan struct{}
	val  V
	err  error
}

// Cache — кэш с загрузкой по промаху и временем жизни записей.
type Cache[K comparable, V any] struct {
	ttl  time.Duration
	now  func() time.Time
	load func(context.Context, K) (V, error)

	mu       sync.Mutex
	items    map[K]cacheEntry[V]
	inflight map[K]*loadCall[V]
}

// NewCache создаёт кэш. now — источник времени (nil — time.Now).
func NewCache[K comparable, V any](ttl time.Duration, now func() time.Time, load func(context.Context, K) (V, error)) *Cache[K, V] {
	if now == nil {
		now = time.Now
	}
	return &Cache[K, V]{
		ttl: ttl, now: now, load: load,
		items:    make(map[K]cacheEntry[V]),
		inflight: make(map[K]*loadCall[V]),
	}
}

// Get возвращает значение ключа.
//   - Запись свежая, пока now() < момент_загрузки + ttl; свежую отдаём
//     без вызова load.
//   - Иначе вызывается load. Все Get этого ключа, пришедшие во время
//     загрузки, ждут её же, а не запускают свою.
//   - Ошибки не кэшируются: следующий Get попробует снова.
//   - Загрузка одного ключа не задерживает Get других ключей.
//   - Отмена ctx прерывает только ожидание этого Get: загрузка
//     доводится до конца, и результат попадает в кэш.
func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	c.mu.Lock()
	if e, ok := c.items[key]; ok && c.now().Before(e.exp) {
		c.mu.Unlock()
		return e.val, nil
	}
	cl, ok := c.inflight[key]
	if !ok {
		cl = &loadCall[V]{done: make(chan struct{})}
		c.inflight[key] = cl
		// Грузим в своей горутине и без отмены вызывающего: его ctx
		// может умереть, а результат нужен остальным.
		go c.run(context.WithoutCancel(ctx), key, cl)
	}
	c.mu.Unlock() // load идёт без замка — другие ключи не ждут

	select {
	case <-cl.done:
		return cl.val, cl.err
	case <-ctx.Done():
		var zero V
		return zero, ctx.Err()
	}
}

func (c *Cache[K, V]) run(ctx context.Context, key K, cl *loadCall[V]) {
	v, err := c.load(ctx, key)
	c.mu.Lock()
	if err == nil {
		c.items[key] = cacheEntry[V]{val: v, exp: c.now().Add(c.ttl)}
	}
	delete(c.inflight, key) // и при ошибке: следующий Get загрузит заново
	cl.val, cl.err = v, err
	c.mu.Unlock()
	close(cl.done)
}
