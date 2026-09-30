package main

import "context"

// Barrier — многоразовый барьер на n участников.
type Barrier struct {
	n int
	// ваши поля
}

func NewBarrier(n int) *Barrier {
	return &Barrier{n: n}
}

// Wait блокируется, пока Wait не вызовут n горутин; тогда все n проходят,
// и барьер сразу готов к следующему поколению.
// Возвращает (true, nil) ровно одной горутине каждого поколения — последней
// пришедшей; остальным (false, nil).
// Если ctx отменён раньше, чем поколение заполнилось, Wait возвращает
// (false, ctx.Err()), и эта горутина больше не считается пришедшей.
func (b *Barrier) Wait(ctx context.Context) (bool, error) {
	// ваш код
	return false, nil
}
