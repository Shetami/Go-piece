package main

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
	// ваши поля
}

func (b *Buf) Write(p []byte) (int, error) {
	// ваш код
	return 0, nil
}

func (b *Buf) WriteString(s string) {
	// ваш код
}

func (b *Buf) String() string {
	// ваш код
	return ""
}

func (b *Buf) Len() int {
	// ваш код
	return 0
}

// Reset очищает буфер.
func (b *Buf) Reset() {
	// ваш код
}
