package main

import (
	"context"
	"time"
)

// Cache — кэш с загрузкой по промаху и временем жизни записей.
type Cache[K comparable, V any] struct {
	// ваши поля
}

// NewCache создаёт кэш. now — источник времени (nil — time.Now).
func NewCache[K comparable, V any](ttl time.Duration, now func() time.Time, load func(context.Context, K) (V, error)) *Cache[K, V] {
	// ваш код
	return &Cache[K, V]{}
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
	// ваш код
	var zero V
	return zero, nil
}
