package main

import (
	"context"
	"sync"
)

// Barrier — многоразовый барьер на n участников.
type Barrier struct {
	n     int
	mu    sync.Mutex
	count int           // сколько пришло в текущее поколение
	gen   chan struct{} // закрывается, когда поколение заполнено
}

func NewBarrier(n int) *Barrier {
	return &Barrier{n: n, gen: make(chan struct{})}
}

// Wait блокируется, пока Wait не вызовут n горутин; тогда все n проходят,
// и барьер сразу готов к следующему поколению.
// Возвращает (true, nil) ровно одной горутине каждого поколения — последней
// пришедшей; остальным (false, nil).
// Если ctx отменён раньше, чем поколение заполнилось, Wait возвращает
// (false, ctx.Err()), и эта горутина больше не считается пришедшей.
func (b *Barrier) Wait(ctx context.Context) (bool, error) {
	b.mu.Lock()
	gen := b.gen // запоминаем СВОЁ поколение
	b.count++
	if b.count == b.n {
		close(gen) // будим всех этого поколения
		b.gen = make(chan struct{})
		b.count = 0
		b.mu.Unlock()
		return true, nil
	}
	b.mu.Unlock()

	select {
	case <-gen:
		return false, nil
	case <-ctx.Done():
		b.mu.Lock()
		defer b.mu.Unlock()
		select {
		case <-gen: // поколение заполнилось одновременно с отменой — мы прошли
			return false, nil
		default:
		}
		b.count-- // уходим: нас больше не ждут
		return false, ctx.Err()
	}
}
