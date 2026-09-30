package main

import "bytes"

// BufferPool переиспользует буферы для сериализации ответов, чтобы на
// каждый запрос не выделять память заново.
type BufferPool struct {
	// ваши поля
}

// NewBufferPool создаёт пул. Буферы, выросшие больше maxCap байт,
// обратно в пул не принимаются — их выбрасывают.
func NewBufferPool(maxCap int) *BufferPool {
	// ваш код
	return &BufferPool{}
}

// Get возвращает пустой буфер (Len() == 0) — новый или из пула.
func (p *BufferPool) Get() *bytes.Buffer {
	// ваш код
	return new(bytes.Buffer)
}

// Put возвращает буфер в пул. Put(nil) ничего не делает.
func (p *BufferPool) Put(b *bytes.Buffer) {
	// ваш код
}

// FormatKV сериализует пары в строку вида "a=1 b=22\n": ключи по
// возрастанию, через пробел, в конце перевод строки; пустая мапа — "\n".
// Использует буфер из пула. Возвращённый слайс принадлежит вызывающему:
// последующие вызовы не должны его менять.
func (p *BufferPool) FormatKV(m map[string]int) []byte {
	// ваш код
	return nil
}
