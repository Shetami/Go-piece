package main

import "unsafe"

// Buf — буфер для склейки байтов в духе strings.Builder.
//
//   - нулевое значение готово к работе;
//   - String не копирует байты (ноль выделений памяти); строка, которую
//     вернул String, не меняется ни при последующих записях, ни после Reset;
//   - копировать Buf после первой записи нельзя: методы записи, вызванные
//     на копии, паникуют (две копии делили бы один массив и портили друг
//     другу данные). Копировать нулевой Buf или Buf сразу после Reset можно.
//   - Len и String на копии работают без паники.
type Buf struct {
	addr *Buf // адрес самого себя, запомненный при первой записи
	buf  []byte
}

func (b *Buf) copyCheck() {
	if b.addr == nil {
		b.addr = b
	} else if b.addr != b {
		// Копия содержит адрес оригинала — значит, её скопировали.
		panic("Buf: запись в копию непустого Buf")
	}
}

func (b *Buf) Write(p []byte) (int, error) {
	b.copyCheck()
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *Buf) WriteString(s string) {
	b.copyCheck()
	b.buf = append(b.buf, s...)
}

func (b *Buf) String() string {
	// Строка смотрит в тот же массив. Это безопасно, потому что байты
	// до len(b.buf) больше никто не меняет: append пишет только дальше,
	// а Reset бросает массив, а не переиспользует его.
	return unsafe.String(unsafe.SliceData(b.buf), len(b.buf))
}

func (b *Buf) Len() int { return len(b.buf) }

// Reset очищает буфер.
func (b *Buf) Reset() {
	// Не b.buf[:0]: следующая запись затёрла бы байты строк,
	// которые уже отдал String.
	b.addr = nil
	b.buf = nil
}
