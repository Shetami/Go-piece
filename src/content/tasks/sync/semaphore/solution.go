package main

import "context"

// Semaphore ограничивает число одновременных владельцев.
type Semaphore struct {
	slots chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	// Буфер на n мест: занять — положить, освободить — забрать.
	return &Semaphore{slots: make(chan struct{}, n)}
}

// Acquire занимает место, ожидая при необходимости.
// Если ctx отменили раньше, возвращает ctx.Err() и места не занимает.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release освобождает место.
func (s *Semaphore) Release() {
	<-s.slots
}
