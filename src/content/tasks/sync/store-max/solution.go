package main

import "sync/atomic"

// StoreMax записывает v в *p, если v больше текущего значения, и сообщает,
// было ли обновление. Безопасна для вызова из многих горутин одновременно.
// Мьютексы не использовать.
func StoreMax(p *atomic.Int64, v int64) bool {
	for {
		cur := p.Load()
		if v <= cur {
			return false
		}
		// CAS запишет v, только если значение всё ещё cur. Если между Load
		// и CAS кто-то успел записать своё — перечитываем и решаем заново.
		if p.CompareAndSwap(cur, v) {
			return true
		}
	}
}
