package main

import (
	"context"
	"errors"
	"slices"
	"sync"
)

var (
	ErrUnknownID = errors.New("unknown id")
	ErrDuplicate = errors.New("duplicate id")
)

// Window — окно неподтверждённых сообщений: отправитель резервирует место
// под сообщение до отправки и освобождает его, когда пришло подтверждение.
// Больше size сообщений «в полёте» быть не может — медленный получатель
// притормаживает отправителя (backpressure). Безопасно для многих горутин.
type Window struct {
	slots chan struct{} // семафор: один элемент в буфере — одно место занято

	mu      sync.Mutex
	pending map[string]struct{}
}

func NewWindow(size int) *Window {
	return &Window{
		slots:   make(chan struct{}, size),
		pending: make(map[string]struct{}),
	}
}

// Acquire резервирует место под сообщение id: ждёт, пока в окне освободится
// место, или отмены ctx (тогда ctx.Err()). Если id уже в окне — ErrDuplicate,
// и занятость окна не меняется.
func (w *Window) Acquire(ctx context.Context, id string) error {
	// Быстрый отказ для дубля, не дожидаясь места.
	w.mu.Lock()
	_, dup := w.pending[id]
	w.mu.Unlock()
	if dup {
		return ErrDuplicate
	}

	// Ждём место без мьютекса: иначе заблокировали бы и Ack,
	// который это место освобождает.
	select {
	case w.slots <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if _, dup := w.pending[id]; dup { // тот же id успела занять другая горутина
		<-w.slots
		return ErrDuplicate
	}
	w.pending[id] = struct{}{}
	return nil
}

// Ack подтверждает id и освобождает его место. Если id нет в окне
// (неизвестный или уже подтверждённый) — ErrUnknownID, окно не меняется.
// Ack никогда не блокируется.
func (w *Window) Ack(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.pending[id]; !ok {
		// Освобождать слот нельзя: повторный ack расширил бы окно,
		// а на пустом окне <-w.slots заблокировался бы навсегда.
		return ErrUnknownID
	}
	delete(w.pending, id)
	<-w.slots // не блокируется: слот для id гарантированно занят
	return nil
}

// InFlight возвращает id сообщений в окне по возрастанию.
func (w *Window) InFlight() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	ids := make([]string, 0, len(w.pending))
	for id := range w.pending {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}
