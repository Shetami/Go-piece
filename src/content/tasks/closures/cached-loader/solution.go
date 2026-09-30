package main

import (
	"sync"
	"time"
)

// Cached оборачивает загрузчик кэшем с временем жизни.
//   - Успешный результат по ключу живёт ttl с момента окончания загрузки:
//     если now() - loadedAt >= ttl, значение устарело и грузится заново.
//   - Ошибки не кэшируются: следующий вызов после ошибки грузит снова.
//   - Одновременные вызовы с одним ключом, которого нет в кэше, делят одну
//     загрузку и получают её результат — и значение, и ошибку.
//   - Разные ключи грузятся параллельно: load не вызывается под общей блокировкой.
//
// Результат безопасен для вызова из многих горутин.
func Cached[K comparable, V any](load func(K) (V, error), ttl time.Duration, now func() time.Time) func(K) (V, error) {
	type entry struct {
		done chan struct{} // закрыт, когда загрузка закончилась
		val  V
		err  error
		at   time.Time
	}
	var (
		mu sync.Mutex
		m  = make(map[K]*entry)
	)
	return func(k K) (V, error) {
		mu.Lock()
		if e, ok := m[k]; ok {
			select {
			case <-e.done:
				// Загружено: val, err, at записаны до close(done) — читать можно.
				if e.err == nil && now().Sub(e.at) < ttl {
					mu.Unlock()
					return e.val, nil
				}
				// Устарело — грузим заново ниже.
			default:
				// Грузится прямо сейчас: ждём ту же загрузку без замка.
				mu.Unlock()
				<-e.done
				return e.val, e.err
			}
		}
		e := &entry{done: make(chan struct{})}
		m[k] = e
		mu.Unlock()

		e.val, e.err = load(k)
		e.at = now()
		if e.err != nil {
			mu.Lock()
			if m[k] == e { // ошибку не кэшируем
				delete(m, k)
			}
			mu.Unlock()
		}
		close(e.done) // будит всех, кто ждал эту загрузку
		return e.val, e.err
	}
}
