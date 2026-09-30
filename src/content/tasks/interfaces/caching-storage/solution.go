package main

import (
	"context"
	"errors"
	"slices"
	"sync"
)

var ErrNotFound = errors.New("not found")

// Storage — медленное хранилище (сеть, диск).
type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, val []byte) error
	Delete(ctx context.Context, key string) error
}

type entry struct {
	val []byte
	err error // не nil — закэшированное «не найдено»
}

// Cache — read-through-декоратор над Storage; сам тоже Storage.
type Cache struct {
	s            Storage
	mu           sync.Mutex
	m            map[string]entry
	hits, misses int
}

// Cached оборачивает s:
//   - Get берёт значение из кэша (hit); иначе идёт в s (miss) и кэширует успех;
//   - «не найдено» (errors.Is(err, ErrNotFound)) тоже кэшируется: повторный Get
//     не ходит в s и снова возвращает ошибку с ErrNotFound;
//   - другие ошибки s не кэшируются;
//   - Put/Delete идут в s; при успехе обновляют кэш (после Delete — это
//     закэшированное «не найдено»); при ошибке s запись кэша удаляется;
//   - наружу и внутрь — только копии: вызывающий, изменивший срез после Put
//     или результат Get, кэш не портит;
//   - пустое значение — тоже значение: Get вернёт его без ошибки;
//   - безопасен для параллельного использования.
func Cached(s Storage) *Cache {
	return &Cache{s: s, m: make(map[string]entry)}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	e, ok := c.m[key] // именно ok: пустое значение отличается от «нет записи»
	if ok {
		c.hits++
		c.mu.Unlock()
		if e.err != nil {
			return nil, e.err
		}
		return slices.Clone(e.val), nil
	}
	c.misses++
	c.mu.Unlock()

	// В хранилище идём без блокировки: медленный вызов не должен держать всех.
	val, err := c.s.Get(ctx, key)
	switch {
	case err == nil:
		c.store(key, entry{val: cloneNonNil(val)})
		return cloneNonNil(val), nil
	case errors.Is(err, ErrNotFound):
		c.store(key, entry{err: err})
		return nil, err
	default:
		return nil, err // временную ошибку не кэшируем
	}
}

func (c *Cache) Put(ctx context.Context, key string, val []byte) error {
	if err := c.s.Put(ctx, key, val); err != nil {
		c.drop(key) // что теперь в s — неизвестно
		return err
	}
	c.store(key, entry{val: cloneNonNil(val)})
	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.s.Delete(ctx, key); err != nil {
		c.drop(key)
		return err
	}
	c.store(key, entry{err: ErrNotFound})
	return nil
}

// Stats — число попаданий и промахов Get.
func (c *Cache) Stats() (hits, misses int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits, c.misses
}

func (c *Cache) store(key string, e entry) {
	c.mu.Lock()
	c.m[key] = e
	c.mu.Unlock()
}

func (c *Cache) drop(key string) {
	c.mu.Lock()
	delete(c.m, key)
	c.mu.Unlock()
}

// cloneNonNil копирует срез; nil превращает в пустой, чтобы «значение есть» не терялось.
func cloneNonNil(b []byte) []byte {
	return append([]byte{}, b...)
}
