package main

import (
	"context"
	"sync"
)

// Var хранит значение (например, текущий конфиг или лидера кластера)
// и оповещает об изменениях любое число читателей без опроса.
type Var[T any] struct {
	mu      sync.Mutex
	val     T
	changed chan struct{} // закрывается при следующем Set
}

func NewVar[T any](v T) *Var[T] {
	return &Var[T]{val: v, changed: make(chan struct{})}
}

// Get возвращает текущее значение и канал, который закроется при
// следующем Set. Значение и канал согласованы: если Set случился после
// Get, канал уже закрыт или закроется.
func (x *Var[T]) Get() (T, <-chan struct{}) {
	// Оба — под одним замком: иначе Set мог бы вклиниться между
	// чтением значения и взятием канала, и изменение потерялось бы.
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.val, x.changed
}

// Set меняет значение и будит всех, кто ждёт на канале из Get.
func (x *Var[T]) Set(v T) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.val = v
	close(x.changed)                // broadcast всем ожидающим
	x.changed = make(chan struct{}) // следующий раунд — новый канал
}

// WaitFor ждёт, пока значение станет удовлетворять pred (текущее тоже
// проверяется), и возвращает его. Если ctx отменён раньше — текущее
// значение и ctx.Err().
func (x *Var[T]) WaitFor(ctx context.Context, pred func(T) bool) (T, error) {
	for {
		v, ch := x.Get()
		if pred(v) {
			return v, nil
		}
		select {
		case <-ch:
		case <-ctx.Done():
			v, _ = x.Get()
			return v, ctx.Err()
		}
	}
}
