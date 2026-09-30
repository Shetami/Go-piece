package main

import (
	"context"
	"time"
)

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
	// ваши поля
}

func NewCache[K comparable, V any](ttl time.Duration, now func() time.Time,
	load func(ctx context.Context, key K) (V, error)) *Cache[K, V] {
	// ваш код
	return &Cache[K, V]{}
}

func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	// ваш код
	var zero V
	return zero, nil
}
