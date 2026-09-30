package main

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

// Storage — медленное хранилище (сеть, диск).
type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, val []byte) error
	Delete(ctx context.Context, key string) error
}

// Cache — read-through-декоратор над Storage; сам тоже Storage.
type Cache struct {
	s Storage
	// ваши поля
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
	return &Cache{s: s}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	// ваш код
	return c.s.Get(ctx, key)
}

func (c *Cache) Put(ctx context.Context, key string, val []byte) error {
	// ваш код
	return c.s.Put(ctx, key, val)
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	// ваш код
	return c.s.Delete(ctx, key)
}

// Stats — число попаданий и промахов Get.
func (c *Cache) Stats() (hits, misses int) {
	// ваш код
	return 0, 0
}
