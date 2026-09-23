package main

import "context"

// Semaphore ограничивает число одновременных владельцев.
type Semaphore struct {
	// ваши поля
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{}
}

// Acquire занимает место, ожидая при необходимости.
// Если ctx отменили раньше, возвращает ctx.Err() и места не занимает.
func (s *Semaphore) Acquire(ctx context.Context) error {
	// ваш код
	return nil
}

// Release освобождает место.
func (s *Semaphore) Release() {
	// ваш код
}
