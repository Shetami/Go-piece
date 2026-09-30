package main

import (
	"bytes"
	"maps"
	"slices"
	"strconv"
	"sync"
)

// BufferPool переиспользует буферы для сериализации ответов, чтобы на
// каждый запрос не выделять память заново.
type BufferPool struct {
	pool   sync.Pool
	maxCap int
}

// NewBufferPool создаёт пул. Буферы, выросшие больше maxCap байт,
// обратно в пул не принимаются — их выбрасывают.
func NewBufferPool(maxCap int) *BufferPool {
	p := &BufferPool{maxCap: maxCap}
	// В пуле лежат указатели: положить туда bytes.Buffer по значению —
	// лишняя аллокация на каждый Put (значение упаковывается в any).
	p.pool.New = func() any { return new(bytes.Buffer) }
	return p
}

// Get возвращает пустой буфер (Len() == 0) — новый или из пула.
func (p *BufferPool) Get() *bytes.Buffer {
	b := p.pool.Get().(*bytes.Buffer)
	b.Reset() // Reset оставляет выделенную память — ради неё пул и нужен
	return b
}

// Put возвращает буфер в пул. Put(nil) ничего не делает.
func (p *BufferPool) Put(b *bytes.Buffer) {
	// Смотрим на Cap, а не на Len: после Reset длина 0, а гигантский
	// массив под буфером остаётся и держал бы память вечно.
	if b == nil || b.Cap() > p.maxCap {
		return
	}
	p.pool.Put(b)
}

// FormatKV сериализует пары в строку вида "a=1 b=22\n": ключи по
// возрастанию, через пробел, в конце перевод строки; пустая мапа — "\n".
// Использует буфер из пула. Возвращённый слайс принадлежит вызывающему:
// последующие вызовы не должны его менять.
func (p *BufferPool) FormatKV(m map[string]int) []byte {
	b := p.Get()
	defer p.Put(b)
	for i, k := range slices.Sorted(maps.Keys(m)) {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.Write(strconv.AppendInt(b.AvailableBuffer(), int64(m[k]), 10))
	}
	b.WriteByte('\n')
	// b.Bytes() смотрит в память буфера, которая после Put уйдёт
	// следующему вызову, — отдаём копию.
	return bytes.Clone(b.Bytes())
}
