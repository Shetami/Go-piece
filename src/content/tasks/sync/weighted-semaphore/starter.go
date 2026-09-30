package main

import (
	"context"
	"errors"
)

// ErrTooLarge — запрос больше ёмкости семафора, он не выполнится никогда.
var ErrTooLarge = errors.New("request exceeds semaphore size")

// Weighted — семафор с весами: задача занимает столько единиц, сколько
// ей нужно (например, мегабайт памяти под обработку файла).
//
// Порядок — FIFO: пока в очереди кто-то ждёт, новый запрос встаёт за ним,
// даже если для нового места хватает. Иначе большой запрос будет вечно
// голодать за потоком маленьких.
type Weighted struct {
	// ваши поля
}

func NewWeighted(size int64) *Weighted {
	// ваш код
	return &Weighted{}
}

// Acquire занимает n единиц, ожидая их освобождения. При отмене ctx
// возвращает ctx.Err() и ничего не занимает. n > size — сразу ErrTooLarge.
func (s *Weighted) Acquire(ctx context.Context, n int64) error {
	// ваш код
	return nil
}

// TryAcquire занимает n единиц без ожидания, если это можно сделать сразу
// и в очереди никого нет.
func (s *Weighted) TryAcquire(n int64) bool {
	// ваш код
	return true
}

// Release возвращает n единиц. Вернуть больше, чем занято, — паника.
func (s *Weighted) Release(n int64) {
	// ваш код
}
