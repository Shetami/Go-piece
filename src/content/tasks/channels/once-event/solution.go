package main

import (
	"context"
	"sync"
)

// Event — одноразовое событие («конфиг загружен», «сервер готов»).
// Пока оно не наступило, ожидающие ждут; после Fire проходят все —
// и те, кто уже ждёт, и те, кто придёт позже. Создаётся через NewEvent.
type Event struct {
	once sync.Once
	done chan struct{}
}

func NewEvent() *Event {
	return &Event{done: make(chan struct{})}
}

// Fire отмечает, что событие наступило. Безопасна при повторном и
// конкурентном вызове: второй и следующие вызовы ничего не делают.
func (e *Event) Fire() {
	// Закрытие канала будит всех получателей сразу, но закрыть можно
	// только один раз — sync.Once защищает от паники при повторе.
	e.once.Do(func() { close(e.done) })
}

// Done возвращает канал, который закрывается при Fire.
// Каждый вызов возвращает один и тот же канал.
func (e *Event) Done() <-chan struct{} {
	return e.done
}

// Wait ждёт события или отмены ctx. nil — событие наступило, иначе ctx.Err().
// Если событие наступило до вызова Wait, возвращает nil даже при отменённом ctx.
func (e *Event) Wait(ctx context.Context) error {
	// select выбирает среди готовых веток случайно, поэтому
	// сначала отдельно проверяем, не наступило ли событие.
	select {
	case <-e.done:
		return nil
	default:
	}
	select {
	case <-e.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
