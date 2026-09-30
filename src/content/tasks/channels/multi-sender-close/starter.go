package main

import (
	"context"
	"errors"
)

var ErrClosed = errors.New("sink closed")

// Sink — канал, в который пишут многие горутины, а закрывает тот, кто
// решил, что хватит (например, при остановке сервиса).
type Sink[T any] struct {
	// ваши поля
}

func NewSink[T any]() *Sink[T] {
	// ваш код
	return &Sink[T]{}
}

// Out — канал для чтения. Каждый вызов — один и тот же канал. Он
// закрывается, когда Close вызван и все начатые Send вернулись.
func (s *Sink[T]) Out() <-chan T {
	// ваш код
	return nil
}

// Send отправляет v, дожидаясь читателя. Возвращает nil, если значение
// принято (и тогда читатель Out его получит), ErrClosed — если Sink
// закрыт до или во время ожидания, ctx.Err() — если отменён ctx.
// Send никогда не паникует.
func (s *Sink[T]) Send(ctx context.Context, v T) error {
	// ваш код
	return nil
}

// Close закрывает Sink: ждущие Send возвращают ErrClosed, новые — тоже.
// Возвращается, когда Out закрыт. Повторный и конкурентный Close безопасны.
func (s *Sink[T]) Close() {
	// ваш код
}
