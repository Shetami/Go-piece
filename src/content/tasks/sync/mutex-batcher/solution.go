package main

import "sync"

// Batcher копит элементы и отдаёт их в flush пачками по size штук —
// например, строки для одного INSERT на много значений. Add вызывают из
// многих горутин одновременно.
//
// flush вызывается без удержания внутренней блокировки Batcher (он может
// идти в базу секундами, и другие Add не должны его ждать) и имеет право
// сохранить полученный слайс у себя. flush никогда не получает пустую пачку.
type Batcher[T any] struct {
	mu    sync.Mutex
	size  int
	buf   []T
	flush func([]T)
}

func NewBatcher[T any](size int, flush func([]T)) *Batcher[T] {
	return &Batcher[T]{size: size, flush: flush, buf: make([]T, 0, size)}
}

// Add добавляет элемент. Если накопилось size элементов, отправляет их
// в flush.
func (b *Batcher[T]) Add(v T) {
	b.mu.Lock()
	b.buf = append(b.buf, v)
	var batch []T
	if len(b.buf) >= b.size {
		batch = b.takeLocked()
	}
	b.mu.Unlock()
	// flush — снаружи мьютекса: медленная запись не держит других писателей.
	if batch != nil {
		b.flush(batch)
	}
}

// Flush немедленно отправляет накопленное (если есть что отправлять).
func (b *Batcher[T]) Flush() {
	b.mu.Lock()
	batch := b.takeLocked()
	b.mu.Unlock()
	if len(batch) > 0 {
		b.flush(batch)
	}
}

// takeLocked забирает накопленное и заводит НОВЫЙ массив. buf[:0] здесь
// нельзя: пачка ушла во flush, и следующие append писали бы поверх неё.
func (b *Batcher[T]) takeLocked() []T {
	if len(b.buf) == 0 {
		return nil
	}
	batch := b.buf
	b.buf = make([]T, 0, b.size)
	return batch
}
