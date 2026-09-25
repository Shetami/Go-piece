package main

import "sync/atomic"

// StoreMax записывает v в *p, если v больше текущего значения, и сообщает,
// было ли обновление. Безопасна для вызова из многих горутин одновременно.
// Мьютексы не использовать.
func StoreMax(p *atomic.Int64, v int64) bool {
	// ваш код
	return false
}
