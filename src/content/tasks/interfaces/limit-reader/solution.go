package main

import "io"

type limited struct {
	r    io.Reader
	left int64 // сколько байт ещё можно отдать
}

func (l *limited) Read(p []byte) (int, error) {
	if l.left <= 0 {
		return 0, io.EOF
	}
	// Не просим у источника больше, чем осталось: иначе прочитанное сверх
	// лимита пришлось бы выбросить, и оно пропало бы для следующего читателя.
	if int64(len(p)) > l.left {
		p = p[:l.left]
	}
	n, err := l.r.Read(p)
	l.left -= int64(n)
	return n, err
}

// Limit возвращает io.Reader, который отдаёт из r не больше n байт,
// а потом возвращает io.EOF — даже если в r данные ещё есть.
func Limit(r io.Reader, n int64) io.Reader {
	// Указатель: Read меняет left, и изменения должны сохраняться между вызовами.
	return &limited{r: r, left: n}
}
