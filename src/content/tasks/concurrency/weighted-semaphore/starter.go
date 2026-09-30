package main

import (
	"context"
	"errors"
)

var ErrTooLarge = errors.New("semaphore: request exceeds size")

// Weighted — семафор с весами на size единиц (как golang.org/x/sync/semaphore).
//
//   - Acquire(ctx, n) занимает n единиц, блокируясь, пока их не хватит.
//     n > size — сразу ErrTooLarge.
//   - Ожидающие обслуживаются строго по очереди (FIFO): пока первый в
//     очереди ждёт, следующие его не обгоняют, даже если им хватило бы.
//   - Если ctx отменён раньше, чем единицы выданы, Acquire возвращает
//     ctx.Err() и ничего не занимает, а очередь двигается дальше.
//   - TryAcquire(n) — без ожидания; false, если единиц мало или есть очередь.
//   - Release(n) возвращает единицы и будит тех, кому теперь хватает.
//     Вернуть больше, чем занято, — паника.
type Weighted struct {
	size int64
	// ваши поля
}

func NewWeighted(size int64) *Weighted { return &Weighted{size: size} }

func (s *Weighted) Acquire(ctx context.Context, n int64) error {
	// ваш код
	return nil
}

func (s *Weighted) TryAcquire(n int64) bool {
	// ваш код
	return false
}

func (s *Weighted) Release(n int64) {
	// ваш код
}
