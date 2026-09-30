package main

import (
	"bytes"
	"sync"
)

// BufferPool — пул *bytes.Buffer поверх sync.Pool для горячего пути
// (сериализация ответов, сборка строк лога). Безопасен для конкурентного
// использования.
type BufferPool struct {
	pool   sync.Pool
	maxCap int
}

// NewBufferPool создаёт пул. Буферы, выросшие больше maxCap байт
// (Cap() > maxCap), обратно в пул не возвращаются.
func NewBufferPool(maxCap int) *BufferPool {
	p := &BufferPool{maxCap: maxCap}
	// New вызывается, когда пул пуст: Get никогда не вернёт nil.
	p.pool.New = func() any { return new(bytes.Buffer) }
	return p
}

// Get возвращает пустой буфер (Len() == 0), никогда не nil.
func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Put возвращает буфер в пул. Put(nil) ничего не делает. Слишком большие
// буферы выбрасываются, чтобы один огромный ответ не остался жить в пуле.
func (p *BufferPool) Put(b *bytes.Buffer) {
	if b == nil || b.Cap() > p.maxCap {
		// Большой буфер отдаём сборщику мусора: иначе пул навсегда
		// держит мегабайты ради редкого большого запроса.
		return
	}
	// Сбрасываем при возврате: следующий Get получит пустой буфер,
	// а память под байты останется.
	b.Reset()
	p.pool.Put(b)
}
