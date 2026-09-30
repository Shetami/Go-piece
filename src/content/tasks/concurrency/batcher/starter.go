package main

import (
	"errors"
	"time"
)

var ErrClosed = errors.New("batcher closed")

// Batcher копит элементы и отдаёт их пачками в flush.
//
// Пачка уходит, когда в ней набралось maxSize элементов или когда с момента
// появления ПЕРВОГО элемента текущей пачки прошло maxWait — что раньше.
// flush вызывается из одной фоновой горутины, последовательно; пачки не
// пустые, порядок элементов сохраняется. flush может сохранить полученный
// слайс у себя — Batcher не должен его потом менять.
type Batcher[T any] struct {
	maxSize int
	maxWait time.Duration
	flush   func([]T)
	// ваши поля
}

func NewBatcher[T any](maxSize int, maxWait time.Duration, flush func([]T)) *Batcher[T] {
	// ваш код
	return &Batcher[T]{maxSize: maxSize, maxWait: maxWait, flush: flush}
}

// Add добавляет элемент. После Close возвращает ErrClosed.
// Безопасен для вызова из разных горутин, в том числе одновременно с Close.
func (b *Batcher[T]) Add(item T) error {
	// ваш код
	return nil
}

// Close отправляет остаток, дожидается последнего flush и останавливает
// фоновую горутину. Повторный вызов ничего не делает.
func (b *Batcher[T]) Close() {
	// ваш код
}
