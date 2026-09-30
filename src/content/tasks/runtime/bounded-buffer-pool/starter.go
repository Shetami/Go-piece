package main

import "bytes"

// BufferPool — пул *bytes.Buffer поверх sync.Pool для горячего пути
// (сериализация ответов, сборка строк лога). Безопасен для конкурентного
// использования.
type BufferPool struct {
	// ваши поля
}

// NewBufferPool создаёт пул. Буферы, выросшие больше maxCap байт
// (Cap() > maxCap), обратно в пул не возвращаются.
func NewBufferPool(maxCap int) *BufferPool {
	// ваш код
	return &BufferPool{}
}

// Get возвращает пустой буфер (Len() == 0), никогда не nil.
func (p *BufferPool) Get() *bytes.Buffer {
	// ваш код
	return nil
}

// Put возвращает буфер в пул. Put(nil) ничего не делает. Слишком большие
// буферы выбрасываются, чтобы один огромный ответ не остался жить в пуле.
func (p *BufferPool) Put(b *bytes.Buffer) {
	// ваш код
}
