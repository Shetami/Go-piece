package main

import (
	"errors"
	"sync"
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

	mu     sync.Mutex // защищает closed и отправку в in
	closed bool
	in     chan T
	done   chan struct{}
}

func NewBatcher[T any](maxSize int, maxWait time.Duration, flush func([]T)) *Batcher[T] {
	b := &Batcher[T]{
		maxSize: maxSize,
		maxWait: maxWait,
		flush:   flush,
		in:      make(chan T),
		done:    make(chan struct{}),
	}
	go b.loop()
	return b
}

// Add добавляет элемент. После Close возвращает ErrClosed.
// Безопасен для вызова из разных горутин, в том числе одновременно с Close.
func (b *Batcher[T]) Add(item T) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return ErrClosed
	}
	b.in <- item // под мьютексом: Close не закроет in посреди отправки
	return nil
}

// Close отправляет остаток, дожидается последнего flush и останавливает
// фоновую горутину. Повторный вызов ничего не делает.
func (b *Batcher[T]) Close() {
	b.mu.Lock()
	if !b.closed {
		b.closed = true
		close(b.in)
	}
	b.mu.Unlock()
	<-b.done
}

func (b *Batcher[T]) loop() {
	defer close(b.done)
	var batch []T
	var timer *time.Timer
	var timeout <-chan time.Time // nil, пока пачка пуста: select его не выбирает

	send := func() {
		if timer != nil {
			timer.Stop()
		}
		timeout = nil
		b.flush(batch)
		batch = nil // новый слайс: старый теперь принадлежит flush
	}

	for {
		select {
		case item, ok := <-b.in:
			if !ok {
				if len(batch) > 0 {
					send()
				}
				return
			}
			if len(batch) == 0 {
				// Отсчёт — от первого элемента пачки, а не от прошлого flush.
				timer = time.NewTimer(b.maxWait)
				timeout = timer.C
			}
			batch = append(batch, item)
			if len(batch) >= b.maxSize {
				send()
			}
		case <-timeout:
			send()
		}
	}
}
