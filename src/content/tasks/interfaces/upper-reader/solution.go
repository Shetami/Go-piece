package main

import "io"

// NewUpperReader возвращает io.Reader, который отдаёт данные r,
// переводя латинские буквы a–z в верхний регистр. Остальное — без изменений.
func NewUpperReader(r io.Reader) io.Reader {
	return upperReader{r}
}

type upperReader struct{ r io.Reader }

func (u upperReader) Read(p []byte) (int, error) {
	n, err := u.r.Read(p)
	// Обрабатываем только p[:n]: остальная часть буфера — мусор.
	// Меняем лишь ASCII-байты: они не встречаются внутри многобайтовых
	// символов UTF-8, поэтому граница чтения посреди буквы не страшна.
	for i := range p[:n] {
		if 'a' <= p[i] && p[i] <= 'z' {
			p[i] -= 'a' - 'A'
		}
	}
	return n, err
}
