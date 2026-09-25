package main

import "io"

// Limit возвращает io.Reader, который отдаёт из r не больше n байт,
// а потом возвращает io.EOF — даже если в r данные ещё есть.
func Limit(r io.Reader, n int64) io.Reader {
	// ваш код
	return r
}
