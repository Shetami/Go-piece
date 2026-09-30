package main

import (
	"context"
	"errors"
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
	// ваши поля
}

func NewWindow(size int) *Window {
	// ваш код
	return &Window{}
}

// Acquire резервирует место под сообщение id: ждёт, пока в окне освободится
// место, или отмены ctx (тогда ctx.Err()). Если id уже в окне — ErrDuplicate,
// и занятость окна не меняется.
func (w *Window) Acquire(ctx context.Context, id string) error {
	// ваш код
	return nil
}

// Ack подтверждает id и освобождает его место. Если id нет в окне
// (неизвестный или уже подтверждённый) — ErrUnknownID, окно не меняется.
// Ack никогда не блокируется.
func (w *Window) Ack(id string) error {
	// ваш код
	return nil
}

// InFlight возвращает id сообщений в окне по возрастанию.
func (w *Window) InFlight() []string {
	// ваш код
	return nil
}
