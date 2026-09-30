package main

import (
	"context"
	"sync"
)

// Version — номер последней применённой версии (например, позиции в
// журнале репликации). Запросы «прочитай свои записи» ждут, пока реплика
// догонит нужную версию.
type Version struct {
	mu   sync.Mutex
	cond *sync.Cond
	n    int64
}

func NewVersion() *Version {
	v := &Version{}
	v.cond = sync.NewCond(&v.mu)
	return v
}

// Set сообщает о новой версии. Версия только растёт: значение меньше
// или равное текущему игнорируется.
func (v *Version) Set(n int64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if n <= v.n {
		return
	}
	v.n = n
	// Broadcast: ждущие ждут разных версий, Signal разбудил бы случайного.
	v.cond.Broadcast()
}

// Get возвращает текущую версию.
func (v *Version) Get() int64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.n
}

// WaitFor блокируется, пока версия не станет >= n, и возвращает nil.
// Если ctx отменён раньше — возвращает ctx.Err(). Если версия уже
// достигнута, возвращает nil сразу, даже при отменённом ctx.
func (v *Version) WaitFor(ctx context.Context, n int64) error {
	// Cond не умеет ждать контекст — будим всех ждущих при отмене.
	// Broadcast под мьютексом: иначе он мог бы проскочить между нашей
	// проверкой ctx.Err() и cond.Wait(), и мы уснули бы навсегда.
	stop := context.AfterFunc(ctx, func() {
		v.mu.Lock()
		v.cond.Broadcast()
		v.mu.Unlock()
	})
	defer stop() // иначе AfterFunc висит до отмены ctx — утечка

	v.mu.Lock()
	defer v.mu.Unlock()
	for v.n < n {
		if err := ctx.Err(); err != nil {
			return err
		}
		v.cond.Wait()
	}
	return nil
}
