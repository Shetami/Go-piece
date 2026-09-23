package main

import "io"

// NewUpperReader возвращает io.Reader, который отдаёт данные r,
// переводя латинские буквы a–z в верхний регистр. Остальное — без изменений.
func NewUpperReader(r io.Reader) io.Reader {
	// ваш код
	return r
}
